# Security policy

## Reporting a vulnerability

Please do not open a public issue for a suspected security vulnerability. Use a private GitHub security advisory or contact the repository maintainers through GitHub.

When reporting, include the affected version, reproduction steps, impact, and a minimal proof of concept where safe. Please allow maintainers reasonable time to investigate and release a fix before public disclosure.

## Privacy model

nowire sends review prompts only to the configured Ollama endpoint. By default that endpoint is local (`127.0.0.1`). Users are responsible for reviewing any custom endpoint before sending source code to it. The GitHub Action should run on a self-hosted runner when source code must remain inside a private network.
