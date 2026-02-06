# Installing Cowardly

This guide covers installing Cowardly from a **release archive** (no Go or repo required). To build from source instead, see the [README Install section](../README.md#install).

## Prerequisites

- **macOS** (Intel or Apple Silicon)
- **Brave Browser** installed at `/Applications/Brave Browser.app`

## Where to download

1. Open the **latest release** on GitHub:

   - **[Releases](https://github.com/miguelmartens/cowardly/releases)** — pick the latest version (e.g. `v0.1.0`).

2. Download the archive that matches your Mac:

   - **Apple Silicon (M1/M2/M3, etc.)** — `cowardly_vX.Y.Z_darwin_arm64.tar.gz`
   - **Intel Mac** — `cowardly_vX.Y.Z_darwin_x86_64.tar.gz`

   Replace `X.Y.Z` with the actual version number. Each archive contains a **directory** (e.g. `cowardly_vX.Y.Z_darwin_arm64/`) with the `cowardly` executable, CHANGELOG.md, LICENSE, and README.md inside it.

## Extract and run

From the directory where you downloaded the file (e.g. `~/Downloads`):

```bash
# Extract (creates a directory, e.g. cowardly_vX.Y.Z_darwin_arm64/, with files inside)
tar xzf cowardly_vX.Y.Z_darwin_arm64.tar.gz

# Enter the directory, make the binary executable, and run
cd cowardly_vX.Y.Z_darwin_arm64
chmod +x cowardly
./cowardly
```

For Intel Macs, use the `darwin_x86_64` archive and the matching directory name:

```bash
tar xzf cowardly_vX.Y.Z_darwin_x86_64.tar.gz
cd cowardly_vX.Y.Z_darwin_x86_64
chmod +x cowardly
./cowardly
```

## Install to your PATH (optional)

To run `cowardly` from anywhere without `./`:

1. **Move the binary** to a directory that’s on your `PATH`, for example:

   - `/usr/local/bin` (often used for user-installed tools; may need `sudo`)
   - `~/bin` (create it and add to `PATH` in your shell config if needed)

2. **Example: install to `/usr/local/bin`**

   ```bash
   tar xzf cowardly_vX.Y.Z_darwin_arm64.tar.gz
   chmod +x cowardly_vX.Y.Z_darwin_arm64/cowardly
   sudo mv cowardly_vX.Y.Z_darwin_arm64/cowardly /usr/local/bin/
   cowardly
   ```

3. **Example: install to `~/bin`**

   ```bash
   mkdir -p ~/bin
   tar xzf cowardly_vX.Y.Z_darwin_arm64.tar.gz
   chmod +x cowardly_vX.Y.Z_darwin_arm64/cowardly
   mv cowardly_vX.Y.Z_darwin_arm64/cowardly ~/bin/
   # Add to PATH if not already (e.g. in ~/.zshrc):  export PATH="$HOME/bin:$PATH"
   cowardly
   ```

## Verify

After installing, run:

```bash
cowardly --help
```

You should see the usage and available flags. Then start the TUI with:

```bash
cowardly
```

Restart **Brave Browser** after applying or resetting settings so changes take effect.

## Upgrading

Download the new release archive, extract it (you’ll get a new versioned directory), then replace your existing `cowardly` binary or run from the new directory. Your config in `~/.config/cowardly/cowardly.yaml` is kept across upgrades.

## See also

- [README — Usage](../README.md#usage) — TUI and CLI options
- [docs/RELEASING.md](RELEASING.md) — How releases are built and named
