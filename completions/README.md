# cc-buddy Shell Completions

This directory contains shell completion scripts for cc-buddy that provide tab completion for commands, options, branch names, and environment names.

## Features

- **Command completion**: Complete `create`, `list`, `delete`, `terminal`
- **Option completion**: Complete all command-line options with `--` prefix
- **Dynamic completion**:
  - Branch names from `git branch -a` for `create` command
  - Environment names from `.cc-buddy/environments.json` for `delete`/`terminal` commands
  - Runtime options (`docker`, `podman`) for `--runtime` flag
  - File/directory paths for `--containerfile` and `--worktree-dir` options

## Installation

### Bash

#### Option 1: Manual Installation
```bash
# Copy the completion script to a standard location
sudo cp completions/cc-buddy-completion.bash /etc/bash_completion.d/cc-buddy

# Or for user-specific installation
mkdir -p ~/.local/share/bash-completion/completions
cp completions/cc-buddy-completion.bash ~/.local/share/bash-completion/completions/cc-buddy
```

#### Option 2: Source in .bashrc
```bash
# Add this line to your ~/.bashrc
source /path/to/cc-buddy/completions/cc-buddy-completion.bash
```

#### Temporary Usage (Current Session Only)
```bash
source completions/cc-buddy-completion.bash
```

### Zsh

#### Option 1: System-wide Installation
```bash
# Copy to system completion directory
sudo cp completions/_cc-buddy /usr/local/share/zsh/site-functions/

# Rebuild completion cache
rm -f ~/.zcompdump && compinit
```

#### Option 2: User-specific Installation
```bash
# Create user completion directory if it doesn't exist
mkdir -p ~/.zsh/completions

# Copy the completion file
cp completions/_cc-buddy ~/.zsh/completions/

# Add to fpath in ~/.zshrc (if not already done)
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
echo 'autoload -U compinit && compinit' >> ~/.zshrc

# Reload your shell or source ~/.zshrc
source ~/.zshrc
```

#### Temporary Usage (Current Session Only)
```bash
# Add to fpath and load
fpath=(./completions $fpath)
autoload -U compinit && compinit
```

### Fish

#### Option 1: User Installation (Recommended)
```bash
# Copy to user completion directory
mkdir -p ~/.config/fish/completions
cp completions/cc-buddy.fish ~/.config/fish/completions/

# Completions will be available in new fish sessions
```

#### Option 2: System-wide Installation
```bash
# Copy to system completion directory (varies by system)
# On most Linux systems:
sudo cp completions/cc-buddy.fish /usr/share/fish/vendor_completions.d/

# On macOS with Homebrew fish:
sudo cp completions/cc-buddy.fish /usr/local/share/fish/vendor_completions.d/
```

#### Temporary Usage (Current Session Only)
```bash
# Source the completion in current fish session
source completions/cc-buddy.fish
```

## Usage Examples

After installation, you can use tab completion with cc-buddy:

```bash
# Complete commands
cc-buddy <TAB>
# Shows: create  delete  list  terminal

# Complete options
cc-buddy create --<TAB>
# Shows: --containerfile  --expose-all  --runtime  --worktree-dir

# Complete branch names for create
cc-buddy create <TAB>
# Shows available git branches (local and remote)

# Complete environment names for delete/terminal
cc-buddy delete <TAB>
# Shows: active environment names from .cc-buddy/environments.json

# Complete runtime options
cc-buddy create my-branch --runtime <TAB>
# Shows: docker  podman
```

## Troubleshooting

### Bash Completions Not Working

1. **Check if bash-completion is installed**:
   ```bash
   # On Ubuntu/Debian
   sudo apt install bash-completion
   
   # On CentOS/RHEL/Fedora
   sudo yum install bash-completion  # or dnf install bash-completion
   
   # On macOS
   brew install bash-completion
   ```

2. **Verify bash-completion is loaded**:
   ```bash
   # Should show bash completion functions
   type _init_completion
   ```

3. **Check if the completion is registered**:
   ```bash
   complete | grep cc-buddy
   ```

### Zsh Completions Not Working

1. **Check if completion system is enabled**:
   ```bash
   # Add to ~/.zshrc if missing
   autoload -U compinit && compinit
   ```

2. **Verify fpath includes completion directory**:
   ```bash
   echo $fpath
   ```

3. **Rebuild completion cache**:
   ```bash
   rm -f ~/.zcompdump*
   compinit
   ```

4. **Check if completion is loaded**:
   ```bash
   which _cc_buddy
   ```

### Fish Completions Not Working

1. **Check if fish is using the completion**:
   ```bash
   # List all completions for cc-buddy
   complete -c cc-buddy
   ```

2. **Verify completion file location**:
   ```bash
   # Check if the file exists in the completion directory
   ls ~/.config/fish/completions/cc-buddy.fish
   ```

3. **Test completion functions manually**:
   ```bash
   # In fish shell, test the helper functions
   __cc_buddy_git_branches
   __cc_buddy_environments
   ```

4. **Restart fish or reload completions**:
   ```bash
   # Restart fish shell or reload configuration
   exec fish
   # Or reload completions specifically
   fish_update_completions
   ```

### General Issues

- **Restart your shell** after installation
- **Check file permissions**: completion files should be readable
- **Verify cc-buddy is in PATH**: completions work best when the command is in your PATH
- **Check for conflicts**: other completion scripts might interfere

## Development

The completion scripts are designed to:

1. **Work offline**: No network calls, uses local git and filesystem data
2. **Handle errors gracefully**: Won't break if git repo is missing or malformed
3. **Be performant**: Minimal overhead for completion generation
4. **Support both absolute and relative paths**: Works with `./cc-buddy` and `cc-buddy`

### Testing

You can test completions manually:

```bash
# For bash (after sourcing the completion)
_cc_buddy_completion
echo "${COMPREPLY[@]}"

# For zsh (after loading completion)
_cc_buddy
```