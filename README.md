# AI Village

AI Village is a small desktop app that runs four focused AI agents in sequence:
Researcher, Planner, Builder, and Reviewer. The Windows version has a native
Windows interface and does not open a terminal window.

## Windows app

1. Download \`AI-Village-windows-x64.exe\` from the latest successful
   **Build Windows App** GitHub Actions run (or from a GitHub Release).
2. Double-click the executable.
3. Paste your NVIDIA API key into the password field.
4. Enter a project goal and click **Run AI Village**.
5. Follow each agent's progress and read the results in the app window.

Create an API key through the NVIDIA API Catalog before using the app. AI
requests are sent to NVIDIA NIM, so an internet connection is required.
The API key is used only in memory for that run and is not saved by the app.
You will need to enter it again the next time you launch AI Village.

The model and endpoint can be changed in the app. Defaults are
\`meta/llama-3.1-70b-instruct\` and
\`https://integrate.api.nvidia.com/v1/chat/completions\`.

## Agent specialties

- **Researcher:** facts, requirements, constraints, and risks
- **Planner:** ordered, achievable steps
- **Builder:** implementation details and code
- **Reviewer:** correctness, risks, and missing checks

Each agent receives a focused system prompt that keeps it within its specialty.
Each agent has a 30-second HTTP timeout.

## Build from source

The Windows desktop app is built with Go and the Walk Windows GUI toolkit.
A Windows build requires Go 1.21 or newer and embeds the application manifest
so the executable is self-contained.

\`\`\`powershell
go mod tidy
go install github.com/akavel/rsrc@latest
rsrc -manifest app.manifest -o rsrc_windows_amd64.syso
go build -trimpath -ldflags="-s -w -H=windowsgui -linkmode=internal" -o AI-Village.exe .
\`\`\`

On non-Windows systems, \`go run . -goal "Your goal"\` remains available as a
command-line version for development. The Windows executable is the GUI app.
