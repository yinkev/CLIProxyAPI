# Security Policy

## Reporting Security Issues

Please do not open public issues for vulnerabilities involving credentials, token leakage, auth bypass, request smuggling, or unsafe local command execution.

Report security concerns privately to the maintainer through GitHub profile contact channels or by opening a minimal private disclosure path if available.

## Sensitive Data

CLIProxyAPI may interact with local CLI auth flows, OAuth-backed accounts, API keys, and provider credentials. Contributors must avoid logging or committing:

- API keys
- OAuth tokens
- Cookies
- Authorization headers
- Local config files containing secrets
- Provider account identifiers not needed for debugging

## Supported Scope

Security reports are most useful when they include:

- Affected endpoint or provider path.
- Minimal reproduction steps.
- Expected vs actual behavior.
- Whether credentials, local files, or command execution are exposed.
- Suggested mitigation, if known.
