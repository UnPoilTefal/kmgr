# kmgr — Kubeconfig Manager

A CLI tool to normalize and manage kubeconfigs with a strict naming convention.  
**Multi-OS:** Linux, macOS, Windows / WSL.

---

## Installation

### Via Homebrew (macOS & Linux)

```bash
brew tap UnPoilTefal/kmgr
brew install kmgr
```

### Via GitHub Releases

Download the latest binary for your platform from [Releases](https://github.com/UnPoilTefal/kmgr/releases):

```bash
# macOS Intel (x86_64)
curl -Lo kmgr https://github.com/UnPoilTefal/kmgr/releases/download/latest/kmgr-darwin-amd64
chmod +x kmgr
sudo mv kmgr /usr/local/bin/

# macOS ARM (Apple Silicon)
curl -Lo kmgr https://github.com/UnPoilTefal/kmgr/releases/download/latest/kmgr-darwin-arm64
chmod +x kmgr
sudo mv kmgr /usr/local/bin/

# Linux x86_64
curl -Lo kmgr https://github.com/UnPoilTefal/kmgr/releases/download/latest/kmgr-linux-amd64
chmod +x kmgr
sudo mv kmgr /usr/local/bin/

# Linux ARM64
curl -Lo kmgr https://github.com/UnPoilTefal/kmgr/releases/download/latest/kmgr-linux-arm64
chmod +x kmgr
sudo mv kmgr /usr/local/bin/
```

### Via `go install`

```bash
go install github.com/UnPoilTefal/kmgr@latest
```

### From source

```bash
git clone https://github.com/UnPoilTefal/kmgr.git
cd kmgr
make install
```

### Verify installation

```bash
kmgr version
```

---

## Quick Start

### 1. Initialize

Run this command once after installation:

```bash
kmgr init
```

This creates the directory structure (`~/.kube/configs/`, `~/.kube/backups/`) and adds to your shell profile (`~/.bashrc` / `~/.zshrc`):

```bash
export KUBECONFIG="${KCFG_DIR:-$HOME/.kube}/config"
source <(kmgr completion bash)  # or zsh
```

Reload your shell:
```bash
source ~/.bashrc   # or ~/.zshrc
```

### 2. Import kubeconfigs

```bash
# From a file
kmgr import -f ~/Downloads/kubeconfig.yaml -u john -c prod-payments

# From clipboard (macOS)
kmgr import --clipboard -u john -c prod-payments

# From pipe (k3d, kind, cloud CLIs)
k3d kubeconfig get mycluster | kmgr import --stdin -u john -c mycluster
kind get kubeconfig --name dev | kmgr import --stdin -u john -c dev
gcloud container clusters get-credentials prod --format=yaml | kmgr import --stdin -u john -c prod

# Overwrite an existing import (auto-backup)
kmgr import -f ~/Downloads/kubeconfig.yaml -u john -c prod-payments --force
```

### 3. Merge & manage

```bash
kmgr list                   # List with current context marked
kmgr use john@prod          # Switch to a context (tab-completion support)
kmgr status                 # Show current context + connectivity test
kmgr merge                  # Merge all sources into ~/.kube/config
```

---

## Naming Convention

kmgr enforces a strict naming convention to keep kubeconfigs organized and predictable:

```
File format    : kubeconfig_{user}@{cluster}.yaml
Context name   : {user}@{cluster}

Example:
  kubeconfig_john@prod-payments.yaml  →  Context: john@prod-payments
```

**One context per file.** Each cluster gets its own file.  
`kmgr merge` combines all sources into a single `~/.kube/config`.

---

## Directory Structure

```
~/.kube/                         # Base directory (configurable via KCFG_DIR)
├── config                       # Active merged kubeconfig (managed by kmgr)
├── configs/                     # Source files (one context each)
│   ├── kubeconfig_john@prod.yaml
│   ├── kubeconfig_john@staging.yaml
│   └── quarantine/              # Non-conformant files (auto-moved)
└── backups/                     # Timestamped backups
```

---

## Environment Variables

| Variable   | Default   | Description |
|-----------|-----------|-------------|
| `KCFG_DIR` | `~/.kube` | Base directory (configs/, backups/, config) |
| `NO_COLOR` | _(unset)_ | Disable ANSI colors ([no-color.org](https://no-color.org)) |

```bash
# Custom directory
export KCFG_DIR="$HOME/.config/kmgr"

# Disable colors
export NO_COLOR=1
```

---

## Commands

| Command | Description |
|---------|-------------|
| `kmgr init` | Initialize structure and shell profile |
| `kmgr import` | Import and normalize a kubeconfig |
| `kmgr merge` | Merge all source files into `~/.kube/config` |
| `kmgr list` | List managed contexts (current marked) |
| `kmgr use <context>` | Switch to a context |
| `kmgr status` | Show current context + connectivity test |
| `kmgr check` | Verify integrity of sources, merged config, and connectivity |
| `kmgr fix` | Auto-fix non-conformant files and permissions |
| `kmgr rename <old> <new>` | Rename a context |
| `kmgr export <context>` | Export context to file or stdout (pipe-ready) |
| `kmgr remove <context>` | Remove a context (with confirmation) |
| `kmgr completion` | Generate shell completion (bash/zsh) |
| `kmgr version` | Show version, config paths, and system info |

---

## Maintenance

### Verify configuration

```bash
kmgr check      # Full system check (sources + merged + connectivity)
kmgr fix        # Auto-fix permissions or move non-conformant files to quarantine
```

### Rename / Export

```bash
kmgr rename john@prod-old john@prod-new

kmgr export john@prod                   # Output to stdout (pipeable)
kmgr export john@prod -f ~/shared.yaml  # Save to file
```

### Global options

```bash
kmgr -q <command>       # Quiet mode: suppress all output except errors
kmgr -h                 # Show help
```

---

## Shell Completion

If `kmgr init` was already run, completion is active after shell reload.  
To manually enable:

```bash
# Bash
source <(kmgr completion bash)

# Zsh
source <(kmgr completion zsh)
```

---

## Multi-platform Support

### Linux & macOS

- Full support including file permissions (0600)
- Clipboard support: `xclip` / `xsel` (X11) or `wl-paste` (Wayland)

### Windows / WSL

- **WSL:** Full native support — Unix permissions respected
- **Windows (native):** ACL-based permission checks (Unix mode 0600 not enforced on Windows NTFS)
- Clipboard support: `powershell.exe Get-Clipboard`

---

## Examples

### Import from file

```bash
kmgr import -f ~/kubeconfig.yaml -u alice -c production
```

### Import from clipboard

```bash
# Copy kubeconfig to clipboard, then:
kmgr import --clipboard -u alice -c staging
```

### Import from cloud provider CLIs

```bash
# k3d
k3d kubeconfig get mycluster | kmgr import --stdin -u alice -c k3d-local

# kind
kind get kubeconfig --name dev | kmgr import --stdin -u alice -c kind-dev

# GKE
gcloud container clusters get-credentials my-cluster --format=yaml | kmgr import --stdin -u alice -c gke-prod

# EKS
aws eks update-kubeconfig --name my-cluster --region us-west-2 --format=yaml | kmgr import --stdin -u alice -c eks-prod
```

### Switch contexts

```bash
kmgr list                           # See all contexts
kmgr use alice@production           # Switch to production
kmgr status                         # Verify connection
```

### Verify system health

```bash
kmgr check      # Full integrity check
kmgr fix        # Auto-fix any issues
```

---

## Features

✅ **Strict naming convention** – Organized, predictable kubeconfig management  
✅ **Merge management** – Single merged config, multiple source files  
✅ **Multi-source import** – File, clipboard, stdin  
✅ **Automatic backup** – Before each merge  
✅ **Secure permissions** – 0600 on all files  
✅ **Connectivity checks** – TCP handshake + TLS + authentication probes  
✅ **Shell completion** – Bash & Zsh support  
✅ **NO_COLOR support** – Respects user preferences  
✅ **Multi-platform** – Linux, macOS, Windows (WSL)  
✅ **Quarantine system** – Non-conformant files auto-moved for safety  

---

## Development

### Build & test

```bash
make build              # Compile to bin/kmgr
make install            # Install to GOPATH/bin
make test               # Run unit tests
make test-v             # Tests with verbose output
make test-race          # Tests with race detector
make check              # Run linters and tests (quick pre-commit check)
make fmt                # Format code
make lint               # Run golangci-lint
make clean              # Remove bin/ directory
```

### Requirements

- **Go:** 1.22+
- **Dependencies:**
  - `github.com/spf13/cobra` (CLI framework)
  - `k8s.io/client-go` (kubeconfig management)

### Zero external runtime dependencies

kmgr compiles to a single standalone binary with no external dependencies required at runtime.

---

## License

MIT License (see [LICENSE](LICENSE) file)

---

## Contributing

Contributions welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines and development setup.

Open an issue or PR on [GitHub](https://github.com/UnPoilTefal/kmgr) with:
- **Bug reports:** Current behavior, expected behavior, reproduction steps
- **Feature requests:** Use case and desired behavior
- **Discussions:** Questions or design proposals
