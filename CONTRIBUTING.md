# Contributing to nowire

Thanks for helping improve private, local code review. Before opening a pull request, run `make fmt`, `make test`, `make race`, `make vet`, and `make build`. Keep changes focused, add regression tests for behavior changes, and update the README when a user-facing flag or workflow changes.

Please do not include real source code, secrets, private repository names, or production data in fixtures. For Ollama behavior, prefer deterministic `httptest` servers over requiring a locally installed model in unit tests.

Pull requests should explain the user problem, the design choice, and how the change was verified. Reviewers will prioritize correctness, privacy, clear failure modes, and compatibility with the standard library-only design.
