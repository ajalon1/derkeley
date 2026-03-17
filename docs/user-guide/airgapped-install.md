# Airgapped installation

If your environment does not have internet access to GitHub or Homebrew, you can install the DataRobot CLI using pre-downloaded release archives.

## Overview

Airgapped installation is useful for:

- **Restricted networks** — environments without direct internet access
- **Internal mirrors** — organizations that mirror releases internally
- **Offline installation** — air-gapped systems with no external connectivity

This guide assumes you have already obtained the CLI release archive for your operating system through an internal file share, USB drive, or local mirror.

## Archive naming convention

DataRobot releases CLI binaries using [GoReleaser](https://goreleaser.com/), which produces archives with standardized names:

### macOS and Linux

```
dr_v{VERSION}_{OS}_{ARCH}.tar.gz
```

Examples:
- `dr_v0.2.37_Darwin_arm64.tar.gz` — macOS with Apple Silicon (M1/M2/M3)
- `dr_v0.2.37_Darwin_x86_64.tar.gz` — macOS with Intel processor
- `dr_v0.2.37_Linux_arm64.tar.gz` — Linux with ARM64 (including Raspberry Pi)
- `dr_v0.2.37_Linux_x86_64.tar.gz` — Linux with x86_64 (Intel/AMD)
- `dr_v0.2.37_Linux_riscv64.tar.gz` — Linux with RISC-V

### Windows

```
dr_v{VERSION}_Windows_x86_64.zip
```

Example:
- `dr_v0.2.37_Windows_x86_64.zip` — Windows with 64-bit processor

## Installation

### macOS and Linux

1. Extract the archive:
   ```bash
   tar -xzf dr_v0.2.37_Darwin_arm64.tar.gz
   ```

2. Create the installation directory:
   ```bash
   mkdir -p ~/.local/bin
   ```

3. Move and make the binary executable:
   ```bash
   mv dr ~/.local/bin/dr && chmod +x ~/.local/bin/dr
   ```

4. Add to PATH:
   ```bash
   echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc && source ~/.bashrc
   ```
   (Use `~/.zshrc` instead of `~/.bashrc` for Zsh.)

5. Verify installation:
   ```bash
   dr --version
   ```

### Windows (PowerShell)

1. Extract the archive:
   ```powershell
   Expand-Archive -Path dr_v0.2.37_Windows_x86_64.zip -DestinationPath .
   ```

2. Create the installation directory:
   ```powershell
   New-Item -ItemType Directory -Path "$env:LOCALAPPDATA\Programs\dr" -Force | Out-Null
   ```

3. Move the binary:
   ```powershell
   Move-Item -Path .\dr.exe -Destination "$env:LOCALAPPDATA\Programs\dr\dr.exe" -Force
   ```

4. Add to PATH:
   ```powershell
   $installDir = "$env:LOCALAPPDATA\Programs\dr"; $userPath = [Environment]::GetEnvironmentVariable("Path", "User"); [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
   ```
   (Restart PowerShell for changes to take effect.)

5. Verify installation:
   ```powershell
   dr --version
   ```

## Optional: Install shell completions

After installing the binary, run:
```bash
dr self completion install --yes
```

For more details, see [Shell completions](shell-completions.md).

## Uninstallation

### macOS and Linux

1. Remove the binary:
   ```bash
   rm ~/.local/bin/dr
   ```

2. Remove from PATH (edit `~/.bashrc` or `~/.zshrc` and remove the line containing `$HOME/.local/bin`):
   ```bash
   sed -i.bak '/\.local\/bin/d' ~/.bashrc
   ```

3. Remove shell completions:
   ```bash
   rm -f ~/.zsh/completions/_dr ~/.oh-my-zsh/custom/completions/_dr ~/.bash_completions/dr /etc/bash_completion.d/dr ~/.zcompdump*
   ```

### Windows (PowerShell)

1. Remove the binary:
   ```powershell
   Remove-Item -Path "$env:LOCALAPPDATA\Programs\dr\dr.exe" -Force
   ```

2. Remove from PATH (restart PowerShell after):
   ```powershell
   $installDir = "$env:LOCALAPPDATA\Programs\dr"; $userPath = [Environment]::GetEnvironmentVariable("Path", "User"); [Environment]::SetEnvironmentVariable("Path", ($userPath -split ';' | Where-Object { $_ -ne $installDir }) -join ';', "User")
   ```

3. Remove shell completions (if installed):
   - Edit your PowerShell profile (`$PROFILE`) and remove any lines referencing `dr` completion.

## Next steps

After installation, you can:

1. **Authenticate** with your DataRobot instance: `dr auth login`
2. **Explore templates**: `dr templates list`
3. **Get help**: `dr --help`

See the [quick start](../../README.md#quick-start) guide for more information.

## Troubleshooting

### Binary not found after installation

Ensure the installation directory is in your PATH:

**macOS/Linux:**
```bash
echo $PATH
```

Should include your installation directory. If not, add it to your shell profile as shown in step 4 above.

**Windows:**
```powershell
[Environment]::GetEnvironmentVariable("Path", "User")
```

Should include your installation directory. If not, follow the PATH update instructions above and restart PowerShell.

### Permission denied (macOS/Linux)

Ensure the binary is executable:

```bash
chmod +x ~/.local/bin/dr
```

### Archive extraction fails

Verify the archive is not corrupted and that you have the correct tools:

**macOS/Linux:**
```bash
tar -tzf dr_v0.2.37_Darwin_arm64.tar.gz
```

**Windows:**
Ensure you're using `Expand-Archive` (built into PowerShell 5.0+) or a compatible extraction tool like 7-Zip.

### Version mismatch

If you're updating from an older version, ensure you're using the correct archive for your architecture and that your old binary has been removed before installing the new one.
