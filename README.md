# AGS AI Village

**A living, explorable headquarters for an AI workforce.**

AGS AI Village is a desktop-first simulation where AI agents have visible identities, roles, workspaces, and activities. The village is the primary interface: dashboards and inspection panels support the world rather than replace it.

This project takes inspiration from the broader idea of a virtual AI organization, while developing its own AGS identity, world, architecture, and interaction model.

## Current prototype

The first world-engine foundation includes:

- A native desktop game window with a hand-built village map.
- Five named residents with distinct roles, personalities, colors, and workplaces.
- Animated movement, ambient activities, a day clock, pause, and click-to-inspect profiles.
- Optional model-driven activity decisions through an OpenAI-compatible chat-completions endpoint.
- Demo behavior when no API key is configured, so the village can be explored without credentials.

The current AI decision loop is intentionally small: residents periodically request one activity and destination. It is a foundation, not a claim that full delegation, persistent memory, or autonomous project execution is already implemented.

## Builds and releases

- Every push to `main` (and the Windows GUI feature branch) runs the **Build Windows App** workflow. Pull requests targeting `main` also build and test.
- Successful build runs upload `AI-Village-windows-x64.exe` as a temporary Actions artifact (kept for 30 days). These builds do **not** create or update GitHub Releases.
- To prepare a public release, open **Actions → Draft Windows Release → Run workflow**, enter a version tag such as `v0.1.0`, and run it. Alternatively, push a `v*` tag.
- The release workflow builds and tests the app, then creates a **draft GitHub Release** with the Windows executable attached. Review it and publish it manually when you're ready.
- Publishing a GitHub Release triggers the **Complete Next Release items** workflow. See [Project release automation](#project-release-automation) for its required one-time setup and approval gate.

## Project release automation

When a GitHub Release is published, the workflow scans the **AI Village Issue Tracker & Developments** GitHub Project (#3) and changes items whose **Status** is exactly **Next Release** to **Done**. Other statuses and other projects are not changed. The workflow paginates through the Project items.

### One-time maintainer setup

The workflow deliberately uses a protected GitHub Environment so a release cannot immediately use the Project-write token without an independent approval.

1. In the repository, open **Settings → Environments** and create an environment named `project-item-updates`.
2. In that environment, configure **Required reviewers** and enable **Prevent self-review** if GitHub offers that setting for the repository. Choose a trusted reviewer who should approve Project-board updates after a release.
3. Create a fine-grained personal access token from GitHub **Settings → Developer settings → Personal access tokens → Fine-grained tokens**. Set the resource owner to the Project owner, grant access to the `AI-Village` repository, and grant **Projects: Read and write** permission. Keep the token private.
4. In **Settings → Environments → project-item-updates → Environment secrets**, add a secret named `AI_VILLAGE_PROJECTS_TOKEN` with that token as its value. Do not add it as a plain repository secret, commit it, or paste it into an issue or chat.
5. Confirm the Project owner, Project number (#3), Status field, and the `Next Release` and `Done` options still match the workflow configuration.

### How it runs

- **Published release:** waits for approval of the `project-item-updates` environment, then updates all matching items.
- **Manual test:** open **Actions → Complete Next Release items → Run workflow**. The `dry_run` input defaults to `true`; this scans and reports matching items without changing them. Only turn it off when an authorized reviewer intends to perform the actual update.
- The same protected environment is required for manual runs. The token is exposed only to the job after the environment's protection rules are satisfied.
- If the token, Project, Status field, or required options are missing, the run fails with an actionable error. The workflow never prints the token.

## Windows desktop build

The **Build Windows App** workflow produces a native `AI-Village-windows-x64.exe` with no console window. Download it from the Artifacts section of a successful [Build Windows App run](https://github.com/FunisWokodeneer/AI-Village/actions/workflows/build-windows-exe.yml), then double-click it. The village opens in demo mode without requiring credentials.

To enable model-driven resident decisions in this prototype, set `AI_VILLAGE_API_KEY` (or `NIM_API_KEY`) and optionally `AI_VILLAGE_ENDPOINT` and `AI_VILLAGE_MODEL` in the Windows environment before launching. A first-run settings screen and secure in-app credential storage are planned; they are not implemented yet.

## Run

Install Go 1.23 or newer, then:

```powershell
go mod tidy
$env:AI_VILLAGE_API_KEY = "your-api-key"
$env:AI_VILLAGE_ENDPOINT = "https://integrate.api.nvidia.com/v1/chat/completions"
$env:AI_VILLAGE_MODEL = "meta/llama-3.1-70b-instruct"
go run .
```

The API key is read from the process environment and is not written to disk. If no key is present, the village runs in demo mode.

Compatible providers can be used by setting the endpoint, model, and key environment variables. Local OpenAI-compatible servers are supported when reachable from the machine.

## Controls

- Click a resident to inspect their role, personality, home, and current activity.
- Press **Space** to pause or resume the simulation.
- Press **Esc** to close the resident profile.

## Architecture direction

The codebase is being developed around three independent systems:

1. **World engine** — rendering, simulation time, movement, navigation, buildings, interaction, and persistence.
2. **Agent engine** — model adapters, identity, reasoning, memory, communication, and delegation.
3. **Work engine** — tools, projects, files, integrations, and task execution.

The long-term design is provider-neutral and extensible. Agent capabilities must be real and observable; visual activity should never be presented as proof that an agent completed work.

## Roadmap

See [ROADMAP.md](ROADMAP.md) for planned milestones and current implementation status.

## Public project

AI Village is an independent AGS Studio project. The reference material is inspiration for the general concept, not a source-code or visual-design template.
