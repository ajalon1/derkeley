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

1. **Locate the correct archive** for your platform and architecture.

   To check your architecture:
   ```bash
   uname -m
   ```

   To check your OS:
   ```bash
   uname -s
   ```

2. **Extract the archive:**

   ```bash
   tar -xzf dr_v0.2.37_Darwin_arm64.tar.gz
   ```

   This creates a `dr` binary in the current directory.

3. **Choose an installation directory** (create if needed):

   Common options:
   - `~/.local/bin` — user-local directory (recommended for non-root users)
   - `/usr/local/bin` — system-wide directory (requires `sudo`)

   Example: create `~/.local/bin` if it doesn't exist:
   ```bash
   mkdir -p ~/.local/bin
   ```

4. **Move the binary and make it executable:**

   ```bash
   mv dr ~/.local/bin/dr
   chmod +x ~/.local/bin/dr
   ```

   (Replace `~/.local/bin` with your chosen directory if different.)

5. **Add to PATH** (if not already present):

   Check if `~/.local/bin` is in your PATH:
   ```bash
   echo $PATH | grep ~/.local/bin
   ```

   If not, add it to your shell profile. For **bash** or **zsh**:

   ```bash
   echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc
   # (or ~/.zshrc for Zsh)
   source ~/.bashrc
   ```

6. **Verify installation:**

   ```bash
   dr --version
   ```

### Windows (PowerShell)

1. **Locate the correct archive** for your Windows version (64-bit).

2. **Extract the archive:**

   ```powershell
   Expand-Archive -Path dr_v0.2.37_Windows_x86_64.zip -DestinationPath .
   ```

   This creates a `dr.exe` binary in the current directory.

3. **Choose an installation directory** (create if needed):

   Recommended: `$env:LOCALAPPDATA\Programs\dr`

   ```powershell
   $installDir = "$env:LOCALAPPDATA\Programs\dr"
   New-Item -ItemType Directory -Path $installDir -Force | Out-Null
   ```

4. **Copy the binary:**

   ```powershell
   Copy-Item -Path .\dr.exe -Destination $installDir\dr.exe -Force
   ```

5. **Add to PATH**:

   ```powershell
   $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
   if ($userPath -notlike "*$installDir*") {
       $newPath = "$userPath;$installDir"
       [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
       $env:Path = $newPath
   }
   ```

   **Note:** You may need to restart PowerShell for PATH changes to take effect.

6. **Verify installation:**

   ```powershell
   dr --version
   ```

## Optional: Install shell completions

Shell completions provide command auto-completion in your terminal. After installing the binary, run:

```bash
dr self completion install --yes
```

This works offline and will install completions for your current shell (Bash, Zsh, or PowerShell).

For more details, see [Shell completions](../user-guide/shell-completions.md).

## Uninstallation

### macOS and Linux

1. **Remove the binary:**

   ```bash
   rm ~/.local/bin/dr
   ```

   (Replace `~/.local/bin` with your installation directory if different.)

2. **Remove from PATH** (if you added it manually):

   Edit your shell profile (`~/.bashrc` or `~/.zshrc`) and remove the line containing `$HOME/.local/bin` or your custom install directory.

3. **Remove shell completions:**

   ```bash
   # Zsh
   rm -f ~/.zsh/completions/_dr
   rm -f ~/.oh-my-zsh/custom/completions/_dr

   # Bash
   rm -f ~/.bash_completions/dr
   rm -f /etc/bash_completion.d/dr

   # Clear Zsh cache
   rm -f ~/.zcompdump*
   ```

### Windows (PowerShell)

1. **Remove the binary:**

   ```powershell
   $installDir = "$env:LOCALAPPDATA\Programs\dr"
   Remove-Item -Path $installDir\dr.exe -Force
   ```

2. **Remove from PATH**:

   ```powershell
   $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
   $newPath = ($userPath -split ';' | Where-Object { $_ -ne $installDir }) -join ';'
   [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
   ```

   **Note:** Restart PowerShell for PATH changes to take effect.

3. **Remove shell completions** (if installed):

   PowerShell completions are typically embedded in your PowerShell profile. Remove any lines referencing `dr` completion if you added them manually.

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

Should include your installation directory. If not, add it to your shell profile as shown in step 5 above.

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
