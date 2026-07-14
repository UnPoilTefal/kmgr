# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2026-07-14

### Added
- **AI mode** (`--ai` flag / `KMGR_AI=1` env): compact, decoration-free output to
  minimize token usage when kmgr is driven by an AI agent. Auto-enabled when the
  `CLAUDECODE` env var is present (Claude Code sessions); `KMGR_AI=0` forces it off. (#17)
- **Mirror sync** (`kmgr sync`): register mirror targets for the merged kubeconfig
  (`sync add`, `sync add --windows`, `sync remove`, `sync status`, `sync now`).
  Mirrors are refreshed automatically after every write of the merged file
  (import, merge, use, rename, remove, fix). Primary use case: publishing the
  kubeconfig to the Windows side (`/mnt/c/Users/<user>/.kube/config`) from WSL2. (#18)

### Changed
- `chmod` failures on the merged file are now a warning instead of a fatal error,
  so `~/.kube/config` may be a symlink to a Windows drvfs mount under WSL2. (#18)

## [0.3.0] and earlier

### Added
- Initial release of kmgr
- Strict kubeconfig naming convention: `kubeconfig_{user}@{cluster}.yaml`
- Multi-platform support (Linux, macOS, Windows)
- Import from file, clipboard, or stdin
- Automatic merge management
- Backup system before modifications
- Shell completion (bash/zsh)
- NO_COLOR support
- Comprehensive CLI with 14 commands

### Features
- **Import**: Import kubeconfigs from multiple sources
- **Merge**: Combine all sources into single config
- **List**: Show managed contexts
- **Use**: Switch contexts with tab completion
- **Status**: Check connectivity and current context
- **Check**: Verify integrity of all sources
- **Fix**: Auto-fix permissions and quarantine invalid files
- **Rename**: Rename contexts
- **Export**: Export contexts to files or stdout
- **Remove**: Remove contexts with confirmation
- **Init**: Initialize directory structure and shell profile
- **Version**: Show version and configuration info
- **Completion**: Generate shell completion scripts

### Technical
- Built with Go 1.22+
- Uses k8s.io/client-go for kubeconfig management
- Zero runtime dependencies (standalone binary)
- Secure permissions (0600) on all files
- Cross-platform clipboard support
- Comprehensive test coverage