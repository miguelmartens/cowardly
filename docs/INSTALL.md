# Installing Cowardly

This guide covers installing Cowardly from a **release archive** (no Go or repo required). To build from source instead, see the [README Install section](../README.md#install).

## Prerequisites

- **macOS** (Intel or Apple Silicon)
- **Brave Browser** installed at `/Applications/Brave Browser.app` (or Brave Beta at `/Applications/Brave Browser Beta.app`; use `--beta` to target it)

## Install with Homebrew

If you use [Homebrew](https://brew.sh/), you can install from the tap (builds from source; Go is installed automatically as a dependency):

```bash
brew tap miguelmartens/cowardly
brew install cowardly
```

Then run `cowardly`. Tap: [github.com/miguelmartens/homebrew-cowardly](https://github.com/miguelmartens/homebrew-cowardly).

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

## If macOS blocks the app (Gatekeeper)

macOS may show: _“cowardly” can’t be opened because Apple cannot check it for malicious software._ The release binary is not code-signed; you can still run it safely:

1. **Right-click → Open** — In Finder, right-click (or Control-click) the `cowardly` binary, choose **Open**, then click **Open** in the dialog. After that, `./cowardly` from the terminal will work.
2. **Open Anyway** — If you already tried to run it, go to **System Settings → Privacy & Security**, scroll down, and click **Open Anyway** next to the message about cowardly.
3. **Remove quarantine** (optional) — Only if you trust the download: `xattr -d com.apple.quarantine cowardly`. You may still need to use Open or Open Anyway once.

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
