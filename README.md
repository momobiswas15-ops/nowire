# nowire

> **Your PR reviewer, offline.** `ollama run` + GitHub Action that roasts your code without sending it to OpenAI.

nowire is a small Go CLI that sends only changed-code context to a local [Ollama](https://ollama.com) server. It returns actionable findings in the terminal, Markdown, JSON, or SARIF. No API keys. No source-code upload. No hosted AI dependency.

## Quick start

```bash
ollama pull qwen2.5-coder:7b
go install github.com/nowire/nowire/cmd/nowire@latest
nowire init
nowire review
```

`nowire review` reviews the working-tree diff. Use `--staged` for staged changes or `nowire review ./path/to/file.go` to scope a review.

## Output and CI

```bash
nowire review --format markdown --output nowire-review.md --fail-on medium
nowire review --format sarif --output nowire.sarif
nowire review --model qwen2.5-coder:14b --ollama http://127.0.0.1:11434
```

Exit codes are **0** for a clean review, **1** when findings meet `--fail-on`, and **2** for an execution error.

### GitHub Action

The included composite action installs Ollama, pulls a model, reviews the diff, and writes the result to the job summary. Add this workflow to a repository:

```yaml
name: local review
on: [pull_request]
permissions:
  contents: read
jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: {fetch-depth: 0}
      - uses: nowire/nowire/action@v0.1.0
        with:
          model: qwen2.5-coder:7b
          fail-on: high
```

For private code, use a self-hosted runner if you do not want source code to leave your infrastructure. The Action itself never calls a hosted AI API.

## Configuration

`nowire init` creates `.nowire.yml` and `.nowire-policy.md`:

```yaml
model: qwen2.5-coder:7b
ollama_url: http://127.0.0.1:11434
fail_on: high
output: terminal
policy: .nowire-policy.md
```

The policy file is injected into every review prompt, so teams can encode conventions such as error handling, logging, security, or testing expectations. Environment variables `NOWIRE_MODEL` and `OLLAMA_HOST` are supported.

## Design

The review pipeline is deliberately boring: collect `git diff`, preserve broad changed context, inject policy, call Ollama's local `/api/generate` endpoint, parse strict JSON, deduplicate findings, sort by file and line, and render the selected output. The implementation uses only the Go standard library, which keeps the binary easy to audit and makes offline installation straightforward.

## Development

```bash
go test ./...
go vet ./...
go build -o nowire ./cmd/nowire
```

## License

MIT
