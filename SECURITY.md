# Security Policy

## Supported Versions

ArchiPulse is currently in pre-alpha. Security fixes are applied to the latest version only.

| Version | Supported |
|---|---|
| latest (main) | ✅ |
| older releases | ❌ |

---

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

If you discover a security vulnerability in ArchiPulse, please report it responsibly by emailing **security@archipulse.org**.

Include in your report:

- A description of the vulnerability and its potential impact
- Steps to reproduce the issue
- Any relevant logs, screenshots, or proof of concept
- Your suggested fix, if you have one

You will receive an acknowledgement within 48 hours. We will keep you informed of the progress toward a fix and may ask for additional information.

---

## Disclosure Policy

- We will confirm the vulnerability and determine its severity
- We will work on a fix and release it as soon as possible
- We will publicly disclose the vulnerability after a fix is available, crediting the reporter unless they prefer to remain anonymous

---

## Scope

The following are in scope for security reports:

- SQL injection or data exposure via the REST API
- Authentication or authorization bypass (when auth is implemented in v1.0)
- Arbitrary file read/write via AOEF/AJX import
- Remote code execution

The following are **out of scope**:

- Vulnerabilities in dependencies (report these to the dependency maintainers)
- Issues in development/test environments
- Denial of service via large file uploads (rate limiting is a roadmap item)

---

## Acknowledgements

We are grateful to the security researchers who help keep ArchiPulse and its users safe.
