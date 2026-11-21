# AI Village

A lightweight cloud-AI command-line project that gives each agent one
non-overlapping specialty: research, planning, implementation, or review.
It is written in Go and uses NVIDIA NIM's OpenAI-compatible API.

## Quick start

1. Create a free NVIDIA NIM API key from the NVIDIA API Catalog.
2. Set the key in Windows PowerShell:
   `$env:NIM_API_KEY="your-nvidia-api-key"`
3. Run the executable:
   `.\ai-village.exe -goal "Create a lightweight task tracker"`

To run from source:

```sh
go run . -goal "Create a lightweight task tracker"
```

The default model is `meta/llama-3.1-70b-instruct`. It and the endpoint can
be changed with `-model` and `-endpoint`. The API key should be supplied with
`NIM_API_KEY`; `-api-key` is available only for temporary manual use.

## Agent specialties

- **Researcher:** facts, requirements, constraints, and risks
- **Planner:** ordered, achievable steps
- **Builder:** implementation details and code
- **Reviewer:** correctness, risks, and missing checks

Each agent receives a focused system prompt that forbids work outside its
specialty. The executable has no external runtime dependencies or locally
downloaded model. All inference happens through NVIDIA NIM, so an internet
connection is required.

## Build

```sh
go test ./...
go build -trimpath -ldflags="-s -w" -o ai-village.exe .
```
