// Package honcho implements a Honcho-specific OpenAI-compatible embedding
// adapter module.
//
// It exposes POST /honcho/v1/embeddings on the proxy server and forwards
// requests to a local OpenAI-compatible embedding server (for example
// llama-server serving Qwen3-Embedding-4B). Honcho's pgvector schema expects
// 1536-dimensional vectors, so the adapter truncates Matryoshka embeddings to
// the configured dimension count and optionally re-applies L2 normalization.
//
// The route is intentionally separate from the normal /v1 routes so that
// Honcho-specific vector truncation can never affect Codex/Gemini/Claude
// traffic flowing through the standard endpoints.
//
// NOTE: this file is a reconstruction of a previously lost untracked module.
// It was rebuilt against the preserved wiring contract in
// clone-local-changes-2026-05-21.patch (config schema, registration call
// sites, hot-reload hook, and documented behavior).
package honcho

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	log "github.com/sirupsen/logrus"
)

// Module is the Honcho embedding adapter route module.
// Upstream removed the shared modules package (Amp); this module registers
// directly on the Gin engine instead of via RouteModuleV2.
type Module struct {
	mu           sync.RWMutex
	cfg          config.HonchoEmbeddingConfig
	registerOnce sync.Once
	client       *http.Client
}

// New creates a new Honcho embedding adapter module.
func New() *Module {
	return &Module{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

// Name returns the module identifier.
func (m *Module) Name() string {
	return "honcho-embedding"
}

// Register wires the adapter route into the Gin engine. Registration is
// idempotent; the route is always attached and the enabled flag is checked
// per-request so the module can be toggled via config hot reload without a
// server restart.
func (m *Module) Register(engine *gin.Engine, cfg *config.Config) error {
	if engine == nil {
		return fmt.Errorf("honcho module: nil gin engine")
	}
	if cfg != nil {
		m.mu.Lock()
		m.cfg = cfg.HonchoEmbedding
		m.mu.Unlock()
	}
	m.registerOnce.Do(func() {
		// Deliberately registered without the API-key middleware: the adapter
		// is a local-only convenience route for a co-located Honcho instance.
		engine.POST("/honcho/v1/embeddings", m.handleEmbeddings)
		log.Debugf("honcho module: registered POST /honcho/v1/embeddings")
	})
	return nil
}

// OnConfigUpdated refreshes the cached adapter configuration on hot reload.
func (m *Module) OnConfigUpdated(cfg *config.Config) error {
	if cfg == nil {
		return nil
	}
	m.mu.Lock()
	m.cfg = cfg.HonchoEmbedding
	m.mu.Unlock()
	log.Debugf("honcho module: config updated (enabled=%v upstream=%q model=%q dims=%d normalize=%v)",
		cfg.HonchoEmbedding.Enabled, cfg.HonchoEmbedding.UpstreamBaseURL,
		cfg.HonchoEmbedding.ServedModel, cfg.HonchoEmbedding.Dimensions, cfg.HonchoEmbedding.Normalize)
	return nil
}

// snapshot returns a copy of the current configuration.
func (m *Module) snapshot() config.HonchoEmbeddingConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

// upstreamEndpoint resolves the upstream embeddings URL from the configured
// base. Both "http://host:port/v1" and "http://host:port" bases are accepted.
func upstreamEndpoint(base string) string {
	b := strings.TrimRight(strings.TrimSpace(base), "/")
	if b == "" {
		return ""
	}
	if strings.HasSuffix(b, "/v1") {
		return b + "/embeddings"
	}
	return b + "/v1/embeddings"
}

// handleEmbeddings proxies an OpenAI-compatible embeddings request to the
// configured upstream, then truncates and re-normalizes the returned vectors.
func (m *Module) handleEmbeddings(c *gin.Context) {
	cfg := m.snapshot()
	if !cfg.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{
			"message": "honcho embedding adapter is disabled (set honcho-embedding.enabled: true)",
			"type":    "honcho_adapter_disabled",
		}})
		return
	}
	endpoint := upstreamEndpoint(cfg.UpstreamBaseURL)
	if endpoint == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{
			"message": "honcho embedding adapter has no upstream-base-url configured",
			"type":    "honcho_adapter_misconfigured",
		}})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "failed to read request body"}})
		return
	}
	var payload map[string]interface{}
	if err = json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "request body is not valid JSON"}})
		return
	}
	// Force the upstream model name and float encoding; truncation operates on
	// float vectors, not base64 blocks.
	if strings.TrimSpace(cfg.ServedModel) != "" {
		payload["model"] = cfg.ServedModel
	}
	payload["encoding_format"] = "float"

	upstreamBody, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "failed to encode upstream request"}})
		return
	}
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, endpoint, bytes.NewReader(upstreamBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "failed to build upstream request"}})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if key := strings.TrimSpace(cfg.UpstreamAPIKey); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{
			"message": fmt.Sprintf("honcho embedding upstream unreachable: %v", err),
			"type":    "honcho_upstream_error",
		}})
		return
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"message": "failed to read upstream response"}})
		return
	}
	if resp.StatusCode != http.StatusOK {
		c.Data(resp.StatusCode, "application/json", respBody)
		return
	}

	var parsed map[string]interface{}
	if err = json.Unmarshal(respBody, &parsed); err != nil {
		c.Data(http.StatusOK, "application/json", respBody)
		return
	}
	adaptEmbeddings(parsed, cfg.Dimensions, cfg.Normalize)
	c.JSON(http.StatusOK, parsed)
}

// adaptEmbeddings truncates every embedding vector in an OpenAI-style
// embeddings response to dims entries and optionally re-applies L2
// normalization (Matryoshka truncation). Vectors shorter than dims are passed
// through unchanged.
func adaptEmbeddings(parsed map[string]interface{}, dims int, normalize bool) {
	if dims <= 0 {
		return
	}
	data, ok := parsed["data"].([]interface{})
	if !ok {
		return
	}
	for _, item := range data {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		raw, ok := entry["embedding"].([]interface{})
		if !ok {
			continue
		}
		vec := make([]float64, 0, len(raw))
		valid := true
		for _, v := range raw {
			f, ok := v.(float64)
			if !ok {
				valid = false
				break
			}
			vec = append(vec, f)
		}
		if !valid {
			continue
		}
		if len(vec) > dims {
			vec = vec[:dims]
		}
		if normalize {
			l2Normalize(vec)
		}
		entry["embedding"] = vec
	}
}

// l2Normalize scales the vector in place to unit length. Zero vectors are
// left untouched.
func l2Normalize(vec []float64) {
	var sum float64
	for _, f := range vec {
		sum += f * f
	}
	norm := math.Sqrt(sum)
	if norm == 0 {
		return
	}
	for i := range vec {
		vec[i] /= norm
	}
}
