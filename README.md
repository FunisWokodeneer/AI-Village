# AI Village

A lightweight local AI command-line project that gives each agent one
non-overlapping specialty: research, planning, implementation, or review.
It is written in Go and uses Ollama so it can run without cloud API keys.

## Quick start

1. Install [Ollama](https://ollama.com/) on Windows.
2. Download a small model suitable for an older laptop:
   `ollama pull qwen2.5:1.5b`
3. Start Ollama, then run the executable:
   `ai-village.exe -goal "Create a lightweight task tracker"`

To run from source:

```sh
go run . -goal "Create a lightweight task tracker"
```

The default model and endpoint can be changed with `-model` and `-endpoint`.

## Agent specialties

- **Researcher:** facts, requirements, constraints, and risks
- **Planner:** ordered, achievable steps
- **Builder:** implementation details and code
- **Reviewer:** correctness, risks, and missing checks

Each agent receives a focused system prompt that forbids work outside its
specialty. The executable has no external runtime dependencies; the Ollama
model is the only optional heavyweight component.

## Build

```sh
go test ./...
go build -trimpath -ldflags="-s -w" -o ai-village.exe .
```
