# Windows Executable Build Workflow

This GitHub Actions workflow automatically builds a Windows `.exe` from your AI-Village repository and publishes it as a downloadable release asset.

## Quick Start

1. **Copy the workflow file** to your repository:
   ```
   .github/workflows/build-windows-exe.yml
   ```

2. **Customize the environment variables** at the top of the workflow:
   ```yaml
   env:
     PROJECT_NAME: "AI-Village"           # Your exe name
     PYTHON_ENTRY_POINT: "main.py"        # Python entry point
     NODE_PACKAGE_NAME: "ai-village"      # Node.js package name
     RUST_BINARY_NAME: "ai-village"       # Rust binary name
     GO_MAIN_PATH: "./cmd/main.go"        # Go main package
   ```

3. **Push a version tag** to trigger a build:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

4. **Or create a GitHub Release** - the workflow triggers on release publication.

5. **Or run manually** from the Actions tab → "Build Windows Executable" → "Run workflow".

## Supported Project Types

The workflow auto-detects your project type:

| Type | Detection | Build Tool | Output |
|------|-----------|------------|--------|
| Python | `pyproject.toml`, `setup.py`, `requirements.txt`, `Pipfile`, `poetry.lock` | PyInstaller | `dist/PROJECT_NAME.exe` |
| Node.js/Electron | `package.json` | electron-builder | `release/build/*.exe` |
| Rust | `Cargo.toml` | cargo | `target/release/BINARY_NAME.exe` |
| Go | `go.mod` | go build | `dist/PROJECT_NAME.exe` |

## Python Projects (PyInstaller)

### Requirements
- Entry point script (default: `main.py`)
- `requirements.txt` or `pyproject.toml` for dependencies

### Customization Options

**Create a custom PyInstaller spec file** (recommended for complex apps):
1. Copy `pyinstaller.spec.template` to your repo root as `AI-Village.spec`
2. Modify the `Analysis` section with your entry point, data files, hidden imports
3. Update workflow to use: `pyinstaller AI-Village.spec`

**Common PyInstaller fixes:**
```yaml
# Add to hiddenimports in spec file:
hiddenimports = [
    'pkg_resources.py2_warn',
    'your_module.submodule',  # Add imports PyInstaller misses
]

# Add data files:
datas = [
    ('assets', 'assets'),
    ('config', 'config'),
    ('src/data', 'data'),
]
```

**For GUI apps** (no console window):
```yaml
console=False  # In spec file or add --windowed flag
```

**For CLI apps** (with console):
```yaml
console=True  # Or remove --windowed flag
```

## Node.js/Electron Projects

### Requirements
- `package.json` with `build` configuration for electron-builder
- Electron as a dependency

### Setup electron-builder

1. Install:
   ```bash
   npm install --save-dev electron-builder
   ```

2. Add to `package.json`:
   ```json
   {
     "build": {
       "appId": "com.yourcompany.ai-village",
       "productName": "AI-Village",
       "win": {
         "target": ["nsis", "portable"],
         "icon": "build/icon.ico"
       },
       "nsis": {
         "oneClick": false,
         "allowToChangeInstallationDirectory": true
       }
     },
     "scripts": {
       "build:win": "electron-builder --win --x64"
     }
   }
   ```

3. Create `build/icon.ico` (256x256 recommended)

4. Copy `electron-builder.config.json.template` to `electron-builder.json` and customize

## Rust Projects

### Requirements
- `Cargo.toml` with a `[[bin]]` section or `cargo build --bin name`

### Configuration
```toml
# Cargo.toml
[[bin]]
name = "ai-village"
path = "src/main.rs"
```

The workflow builds with `cargo build --release --bin ai-village`

## Go Projects

### Requirements
- `go.mod` file
- Main package at `./cmd/main.go` (customize `GO_MAIN_PATH`)

### Configuration
The workflow builds with:
```bash
go build -ldflags "-s -w -X main.version=v1.0.0" -o dist/AI-Village.exe ./cmd/main.go
```

Add version variable to your main:
```go
package main

var version = "dev"

func main() {
    fmt.Printf("AI-Village v%s\n", version)
}
```

## Artifacts & Releases

### Workflow Outputs
- **GitHub Actions Artifact**: Downloadable from Actions tab (30-day retention)
- **GitHub Release Asset**: Attached to release (permanent)

### Triggering a Release Build

**Option 1: Push a tag**
```bash
git tag v1.2.3
git push origin v1.2.3
```

**Option 2: Create Release on GitHub**
1. Go to Releases → "Create a new release"
2. Choose tag (or create new)
3. Click "Publish release"

**Option 3: Manual trigger**
1. Go to Actions → "Build Windows Executable"
2. Click "Run workflow"
3. Select branch/tag
4. Click "Run workflow"

## Troubleshooting

### Build Fails: "No .exe file found"
- Check the build logs for the actual output path
- Add the path to the `paths` array in the "Determine artifact path" step

### PyInstaller: ModuleNotFoundError
- Add missing imports to `hiddenimports` in spec file
- Use `--collect-all package_name` for packages with data files

### PyInstaller: Executable too large
- Add more excludes to spec file
- Use `--exclude-module` for unused dependencies
- Consider UPX compression (enabled by default)

### Electron: Code signing errors
- Set `verifyUpdateCodeSignature: false` in build config
- For production, configure code signing certificates

### Rust: Linker errors
- Ensure `x86_64-pc-windows-msvc` target is installed
- Visual Studio Build Tools required on Windows runners

### Go: Version not embedded
- Ensure `main.version` variable exists in your main package
- Check ldflags syntax: `-X main.version=v1.0.0`

## Customizing for AI-Village

### If AI-Village is a Python project:
1. Ensure `main.py` (or your entry point) exists
2. List all dependencies in `requirements.txt`
3. Add any data directories to the `--add-data` flags in the workflow
4. Test locally: `pyinstaller --onefile --windowed main.py`

### If AI-Village uses a different structure:
- Modify the `PYTHON_ENTRY_POINT` env var
- Add custom build steps in the workflow before/after the build step
- For monorepos, add a `working-directory` to the job

## Security Notes

- The workflow uses `GITHUB_TOKEN` (no secrets needed for public repos)
- For private repos, ensure Actions have "Read and write permissions"
- Code signing requires additional secrets (not included in this template)

## File Structure

```
.github/
└── workflows/
    ├── build-windows-exe.yml          # Main workflow (copy this)
    ├── pyinstaller.spec.template      # Python spec template
    ├── electron-builder.config.json.template  # Electron config template
    └── README-WINDOWS-BUILD.md        # This file
```

## Testing Locally

Before pushing, test your build locally:

**Python:**
```bash
pip install pyinstaller
pyinstaller --onefile --windowed main.py
dist/main.exe --version
```

**Node.js:**
```bash
npm run build:win
ls release/build/
```

**Rust:**
```bash
cargo build --release
target/release/ai-village.exe --version
```

**Go:**
```bash
go build -o ai-village.exe ./cmd/main.go
./ai-village.exe --version
```

## Support

- GitHub Actions docs: https://docs.github.com/en/actions
- PyInstaller docs: https://pyinstaller.org/
- electron-builder docs: https://www.electron.build/
- Rust cargo docs: https://doc.rust-lang.org/cargo/
- Go build docs: https://pkg.go.dev/cmd/go