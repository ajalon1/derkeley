# Cloud Template Quick Start

Get a DataRobot CLI Cloud Template running in 2 minutes.

## Create Template (in Factory UI)

1. Go to **Settings → Cloud Templates**
2. Click **Create Template**
3. Fill in:
   - **Repository:** DataRobot CLI repo URL
   - **Name:** e.g., `dr-main`
   - **Setup Script:** Copy one of the options below
4. Click **Create** and wait ~1-5 minutes

## Setup Script Options

### Option A: Full Setup (Complete validation)
```bash
#!/usr/bin/env bash
set -euo pipefail

echo "🚀 Setting up DataRobot CLI Cloud Template..."

if ! command -v task &> /dev/null; then
    echo "📦 Installing Task runner..."
    TASK_INSTALL_DIR="${TASK_INSTALL_DIR:-.local/bin}"
    mkdir -p "$TASK_INSTALL_DIR"
    sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b "$TASK_INSTALL_DIR"
    export PATH="$TASK_INSTALL_DIR:$PATH"
fi

echo "🔧 Initializing development environment..."
task dev-init

echo "🧹 Running linters and formatters..."
task lint

echo "🔨 Building CLI binary..."
task build

echo "✅ Running tests..."
task test

echo "✨ Cloud Template setup complete!"
```
⏱️ **Time:** 3-5 minutes

### Option B: Fast Setup (Essentials only)
```bash
#!/usr/bin/env bash
set -euo pipefail

echo "🚀 Setting up DataRobot CLI Cloud Template (minimal)..."

if ! command -v task &> /dev/null; then
    echo "📦 Installing Task runner..."
    TASK_INSTALL_DIR="${TASK_INSTALL_DIR:-.local/bin}"
    mkdir -p "$TASK_INSTALL_DIR"
    sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b "$TASK_INSTALL_DIR"
    export PATH="$TASK_INSTALL_DIR:$PATH"
fi

echo "🔧 Initializing development environment..."
task dev-init

echo "🔨 Building CLI binary..."
task build

echo "✨ Cloud Template setup complete!"
```
⏱️ **Time:** 1-2 minutes

## Use Template (in Factory Session)

1. Start a Factory session
2. Click **Machine Connection** (session start page)
3. Select **Remote** tab
4. Pick your template (e.g., `dr-main`)
5. Click **Connect**

Green indicator appears → You're connected!

## Common Commands

```bash
# Run the CLI
task run

# Build binary
task build

# Run tests
task test

# Format & lint code
task lint

# Run CLI with args
task run -- auth check
task run -- version
```

## Tips

- **Full setup:** Best for sharing with team (validates everything)
- **Fast setup:** Best for quick iterations (skip linting/tests initially)
- **Run tests later:** `task test` after making changes
- **Check coverage:** `task test-coverage` opens HTML report

## Troubleshooting

| Problem | Fix |
| ------- | --- |
| "Setup script failed" | Check build logs in Factory UI                   |
| Task not found        | Restart terminal or re-connect                   |
| Slow build            | Use Fast Setup option, add tests later           |
| Permissions error     | Scripts use `.local/bin` which doesn't need sudo |

## Next Steps

1. Create template with Option A (full) or Option B (fast)
2. Connect from a session
3. Run `task run -- --help` to verify
4. Share the template URL with your team

📚 **Full Guide:** See `CLOUD_TEMPLATE_GUIDE.md` for detailed docs
📖 **More Info:** See `README.md` for customization & best practices