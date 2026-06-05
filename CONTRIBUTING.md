# Contributing to CLIProxyAPI

Thanks for helping improve CLIProxyAPI.

This project focuses on API compatibility, CLI-backed provider workflows, streaming behavior, tool/function-call translation, and practical AI developer tooling.

## Good First Contributions

Good PRs are usually small and focused:

- Fix a protocol compatibility bug.
- Add or improve tests for streaming/tool-call behavior.
- Improve docs for a provider, CLI, or client integration.
- Add validation around config or auth edge cases.
- Improve examples without changing runtime behavior.

## Development Workflow

1. Fork the repository.
2. Create a focused branch.
3. Make the smallest change that solves the problem.
4. Run relevant checks:

```bash
go test ./...
go build ./cmd/server
```

5. Update docs when behavior or setup changes.
6. Open a PR with reproduction steps and validation output.

## PR Expectations

Please include:

- What changed.
- Why it changed.
- How you tested it.
- Any compatibility risks.
- Related issue or upstream behavior, if applicable.

## Compatibility Guidelines

When changing request/response translation:

- Preserve existing provider behavior unless the current behavior is clearly wrong.
- Add tests for new tool/function-call formats.
- Prefer explicit protocol mapping over ad hoc string manipulation.
- Keep provider-specific behavior isolated where possible.

## Security

Do not commit secrets, OAuth tokens, API keys, cookies, or local credential files. Use `.env.example` and documented config placeholders.
