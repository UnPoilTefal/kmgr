# Homebrew Installation Instructions

kmgr can be installed via Homebrew on macOS and Linux.

## Prerequisites

- Homebrew installed ([brew.sh](https://brew.sh/))
- For Linux: Linuxbrew or regular Homebrew

## Installation

### Option 1: Official Tap (when available)

```bash
# Add the tap (replace with actual tap once created)
brew tap UnPoilTefal/kmgr

# Install kmgr
brew install kmgr
```

### Option 2: Direct Formula Installation

```bash
# Download the formula
curl -O https://raw.githubusercontent.com/UnPoilTefal/kmgr/main/Formula/kmgr.rb

# Install from local formula
brew install --formula ./kmgr.rb
```

### Option 3: Manual Installation

```bash
# Download the appropriate binary for your platform from GitHub releases
# Example for macOS Intel:
curl -Lo kmgr https://github.com/UnPoilTefal/kmgr/releases/download/v0.1.0/kmgr-darwin-amd64
chmod +x kmgr
sudo mv kmgr /usr/local/bin/

# Or for Linux:
curl -Lo kmgr https://github.com/UnPoilTefal/kmgr/releases/download/v0.1.0/kmgr-linux-amd64
chmod +x kmgr
sudo mv kmgr /usr/local/bin/
```

## Verification

```bash
kmgr version
kmgr --help
```

## Updating

```bash
# If installed via tap
brew update
brew upgrade kmgr

# If installed manually, download new version from releases
```

## Creating a Homebrew Tap

To create an official Homebrew tap:

1. Create a new GitHub repository: `UnPoilTefal/homebrew-kmgr`
2. Copy the Formula/kmgr.rb file to the repository
3. Update the formula with correct URLs and checksums
4. Users can then install with: `brew install UnPoilTefal/kmgr/kmgr`

## Formula Maintenance

When releasing new versions:

1. Update the Formula/kmgr.rb with new version and SHA256
2. Test the formula locally: `brew install --build-from-source ./kmgr.rb`
3. Commit and push changes
4. Update this documentation if needed