# Updating CLIProxyAPI from Upstream

This guide documents how to update your fork from the upstream repository while preserving local changes.

## Repository Setup

```
origin    -> https://github.com/yinkev/CLIProxyAPI.git (your fork)
upstream  -> https://github.com/router-for-me/CLIProxyAPI.git (main repo)
```

## Update Process

### 1. Fetch Latest Tags and Commits

```bash
git fetch --tags upstream
```

### 2. Check for Local Changes

```bash
git status
git diff --stat
```

### 3. Stash Local Changes (if any)

If you have uncommitted changes that would conflict with the merge:

```bash
git stash push -m "Local changes before update"
```

### 4. Merge Upstream

```bash
git merge upstream/main --no-edit
```

### 5. Reapply Your Changes

```bash
git stash pop
```

If there are conflicts, resolve them manually, then:

```bash
git add <resolved-files>
```

### 6. Verify the Update

```bash
git log --oneline -3
git describe --tags --abbrev=0
```

## Example: Update to v6.7.9

```bash
# Fetch latest
git fetch --tags upstream

# Stash local work
git stash push -m "Local changes before v6.7.9 update"

# Merge upstream
git merge upstream/main --no-edit

# Reapply local changes
git stash pop

# Verify
git describe --tags --abbrev=0  # Should show v6.7.9
```

## What We Updated (v6.7.9)

**Date:** January 18, 2026

**Changes merged:**
- 36 files modified
- 2,694 lines added / 660 lines removed

**Key features:**
- Reasoning state tracking and summary handling improvements
- Gemini API compatibility fix (enum values to strings)
- Thinking validation with `fromSuffix` parameter and budget logic
- Token filename prefix refactoring for Qwen/iFlow
- README updates (ZeroLimit project added)

**Local changes preserved:**
- `internal/api/handlers/management/auth_files.go`
- `internal/api/server.go`
- `internal/registry/model_definitions.go`

---

## Contributing Back to Upstream

**IMPORTANT: Do NOT create a PR from your fork's main branch!**

Your fork's `main` branch contains your entire commit history, including local customizations. If you PR from there, you'll submit ALL your changes, not just the feature you want to contribute.

### Correct Workflow: Cherry-Pick

```bash
# 1. Create a clean branch from upstream
git fetch upstream
git checkout -b feat/my-feature upstream/main

# 2. Make your changes and commit
# ... edit files ...
git add <files>
git commit -m "feat: add my feature"

# 3. Push to your fork
git push -u origin feat/my-feature

# 4. Create PR from this clean branch
gh pr create --repo router-for-me/CLIProxyAPI
```

### If You Already Committed to Your Fork's Main

```bash
# 1. Note your commit hash
git log --oneline -5  # Find your commit, e.g., abc1234

# 2. Create clean branch from upstream
git checkout -b feat/my-feature-clean upstream/main

# 3. Cherry-pick ONLY your specific commit
git cherry-pick abc1234

# 4. Push and create PR
git push -u origin feat/my-feature-clean
gh pr create --repo router-for-me/CLIProxyAPI
```

### What We Learned (v6.7.31)

**Bad:** Created PR #1312 from fork branch → 1824 additions, 14 files (included entire fork history)

**Good:** Created PR #1313 from cherry-picked branch → 38 additions, 1 file (only our feature)

---

## PRs Submitted

| PR | Feature | Status |
|----|---------|--------|
| #1313 | `GetAllStaticModels()` + `LookupStaticModelInfo` consistency | Pending review |
| #1317 | `code_execution` + `url_context` tool passthrough | Pending review |

## New Features Added (v6.7.31+)

### Gemini Tools Passthrough

Added support for Gemini's built-in tools:

```json
{
  "tools": [
    {"google_search": {}},
    {"code_execution": {}},
    {"url_context": {}}
  ]
}
```

- **google_search** - Grounding with Google Search (already in upstream)
- **code_execution** - Agentic Vision, Python execution (Issue #1318)
- **url_context** - Live URL content fetching (Issue #1318)

---

## Using Agentic Vision

Gemini 3 Flash's **Agentic Vision** uses code execution to analyze images with Python. The model can draw bounding boxes, annotations, count objects, and manipulate images.

### Requirements

1. Model: `gemini-3-flash-preview` or `gemini-2.5-flash`
2. Tool: `code_execution` enabled
3. Input: An image (base64 or URL)

### Basic Request

```bash
curl http://localhost:8317/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-proxy" \
  -d '{
    "model": "gemini-3-flash-preview",
    "messages": [{
      "role": "user",
      "content": [
        {"type": "text", "text": "Count the objects and draw bounding boxes around each one."},
        {"type": "image_url", "image_url": {"url": "data:image/png;base64,YOUR_BASE64_HERE"}}
      ]
    }],
    "tools": [{"code_execution": {}}],
    "max_tokens": 4000
  }'
```

### Example Use Cases

#### 1. Object Detection with Bounding Boxes
```json
{
  "content": [
    {"type": "text", "text": "Identify all objects and draw bounding boxes with labels."},
    {"type": "image_url", "image_url": {"url": "data:image/png;base64,..."}}
  ],
  "tools": [{"code_execution": {}}]
}
```

#### 2. Image Occlusion for Studying
```json
{
  "content": [
    {"type": "text", "text": "Create image occlusion by drawing red boxes over key terms and answers for studying."},
    {"type": "image_url", "image_url": {"url": "data:image/png;base64,..."}}
  ],
  "tools": [{"code_execution": {}}]
}
```

#### 3. Chart/Graph Analysis
```json
{
  "content": [
    {"type": "text", "text": "Extract data from this chart using code execution."},
    {"type": "image_url", "image_url": {"url": "data:image/png;base64,..."}}
  ],
  "tools": [{"code_execution": {}}]
}
```

### Response Format

The model returns Python code that would annotate/analyze the image:

```python
import PIL.Image
import PIL.ImageDraw

img = PIL.Image.open('input_file_0.png')
draw = PIL.ImageDraw.Draw(img)

# Bounding boxes with normalized coordinates [ymin, xmin, ymax, xmax]
boxes = [
    {'box_2d': [100, 80, 200, 300], 'label': 'object1'},
    {'box_2d': [250, 150, 400, 350], 'label': 'object2'},
]

for box in boxes:
    ymin, xmin, ymax, xmax = box['box_2d']
    # Convert normalized (0-1000) to pixel coordinates
    left = xmin * width / 1000
    top = ymin * height / 1000
    right = xmax * width / 1000
    bottom = ymax * height / 1000
    draw.rectangle([left, top, right, bottom], outline='red', width=3)

img.save('transformed_image.png')
```

### Available Python Libraries

- `PIL` (Pillow) - Image manipulation
- `numpy` - Numerical operations
- `pandas` - Data analysis
- `matplotlib` - Plotting

### Limitations

- 30 second timeout per code execution
- Up to 5 executions per request
- No file I/O beyond input/output images
- No custom library installation

### Combining Tools

You can use multiple tools together:

```json
{
  "tools": [
    {"code_execution": {}},
    {"google_search": {}},
    {"url_context": {}}
  ]
}
```

### URL Context Example

Fetch and analyze web content:

```json
{
  "messages": [{
    "role": "user",
    "content": "Summarize the content at https://example.com/article"
  }],
  "tools": [{"url_context": {}}]
}
```
