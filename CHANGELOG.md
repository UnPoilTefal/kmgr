# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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