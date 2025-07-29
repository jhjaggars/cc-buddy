# Fish completion for cc-buddy
# To install: copy this file to ~/.config/fish/completions/

# Function to get git branches
function __cc_buddy_git_branches
    if git rev-parse --git-dir >/dev/null 2>&1
        git branch -a 2>/dev/null | sed 's/^[* ] //' | sed 's/^remotes\///' | grep -v '^HEAD' | sort -u
    end
end

# Function to get environment names
function __cc_buddy_environments
    set -l metadata_file ".cc-buddy/environments.json"
    if test -f "$metadata_file"
        jq -r '.environments[].name' "$metadata_file" 2>/dev/null
    end
end

# Function to check if we're completing the first argument after a command
function __cc_buddy_needs_arg
    set -l cmd (commandline -opc)
    set -l last_arg $cmd[-1]
    
    # Check if the last argument is a command that needs an argument
    switch $last_arg
        case create delete terminal
            return 0
        case '*'
            return 1
    end
end

# Function to check which command we're completing for
function __cc_buddy_current_command
    set -l cmd (commandline -opc)
    for arg in $cmd
        switch $arg
            case create list delete terminal
                echo $arg
                return
        end
    end
end

# Remove any existing completions for cc-buddy
complete -c cc-buddy -e

# Global options (available for all commands)
complete -c cc-buddy -l worktree-dir -d "Set worktree directory" -r -F
complete -c cc-buddy -l containerfile -d "Specify containerfile path" -r -F
complete -c cc-buddy -l runtime -d "Override container runtime" -x -a "docker podman"
complete -c cc-buddy -l expose-all -d "Publish all container ports to host"
complete -c cc-buddy -l force -d "Force overwrite existing files"
complete -c cc-buddy -s t -l terminal -d "Launch terminal after creation"
complete -c cc-buddy -s h -l help -d "Show help message"
complete -c cc-buddy -s v -l version -d "Show version information"

# Commands (only show if no command has been specified yet)
complete -c cc-buddy -n "not __fish_seen_subcommand_from init create list delete terminal" -a "init" -d "Create Containerfile.dev in current directory"
complete -c cc-buddy -n "not __fish_seen_subcommand_from init create list delete terminal" -a "create" -d "Create new development environment"
complete -c cc-buddy -n "not __fish_seen_subcommand_from init create list delete terminal" -a "list" -d "List all active environments"
complete -c cc-buddy -n "not __fish_seen_subcommand_from init create list delete terminal" -a "delete" -d "Delete development environment"
complete -c cc-buddy -n "not __fish_seen_subcommand_from init create list delete terminal" -a "terminal" -d "Open shell in running environment"

# Completions for create command
complete -c cc-buddy -n "__fish_seen_subcommand_from create; and __cc_buddy_needs_arg" -a "(__cc_buddy_git_branches)" -d "Git branch"

# Completions for delete command
complete -c cc-buddy -n "__fish_seen_subcommand_from delete; and __cc_buddy_needs_arg" -a "(__cc_buddy_environments)" -d "Environment name"

# Completions for terminal command
complete -c cc-buddy -n "__fish_seen_subcommand_from terminal; and __cc_buddy_needs_arg" -a "(__cc_buddy_environments)" -d "Environment name"

# No additional arguments for list command (only global options apply)