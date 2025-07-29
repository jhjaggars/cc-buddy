#!/bin/bash

# Bash completion for cc-buddy
# To install: source this file in your .bashrc or copy to /etc/bash_completion.d/

_cc_buddy_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Main commands
    local commands="init create list delete terminal"
    
    # Global options
    local global_opts="--worktree-dir --containerfile --runtime --expose-all --force --terminal -t -h --help -v --version"

    # Function to get git branches for create command
    _get_git_branches() {
        if git rev-parse --git-dir >/dev/null 2>&1; then
            # Get local and remote branches, clean up the output
            git branch -a 2>/dev/null | sed 's/^[* ] //' | sed 's/^remotes\///' | grep -v '^HEAD' | sort -u
        fi
    }

    # Function to get environment names from cc-buddy metadata
    _get_environments() {
        local metadata_file=".cc-buddy/environments.json"
        if [[ -f "$metadata_file" ]]; then
            # Extract environment names from JSON
            jq -r '.environments[].name' "$metadata_file" 2>/dev/null || true
        fi
    }

    # Handle options that take arguments
    case "$prev" in
        --worktree-dir)
            # Complete directory paths
            COMPREPLY=($(compgen -d -- "$cur"))
            return 0
            ;;
        --containerfile)
            # Complete file paths
            COMPREPLY=($(compgen -f -- "$cur"))
            return 0
            ;;
        --runtime)
            # Complete with docker or podman
            COMPREPLY=($(compgen -W "docker podman" -- "$cur"))
            return 0
            ;;
    esac

    # If we're at the first argument position (after cc-buddy)
    if [[ $COMP_CWORD -eq 1 ]]; then
        COMPREPLY=($(compgen -W "$commands $global_opts" -- "$cur"))
        return 0
    fi

    # Find the command in the argument list
    local command=""
    local i=1
    while [[ $i -lt $COMP_CWORD ]]; do
        case "${COMP_WORDS[$i]}" in
            init|create|list|delete|terminal)
                command="${COMP_WORDS[$i]}"
                break
                ;;
        esac
        ((i++))
    done

    # Handle completion based on the command
    case "$command" in
        init)
            # For init command, only complete with options
            local init_opts="--force"
            COMPREPLY=($(compgen -W "$init_opts" -- "$cur"))
            ;;
        create)
            # For create command, complete with branch names and options
            local create_opts="--worktree-dir --containerfile --runtime --expose-all"
            if [[ "$cur" == -* ]]; then
                COMPREPLY=($(compgen -W "$create_opts" -- "$cur"))
            else
                # Check if we already have a branch name
                local has_branch=false
                local j=$((i+1))
                while [[ $j -lt $COMP_CWORD ]]; do
                    if [[ "${COMP_WORDS[$j]}" != -* && "${COMP_WORDS[$j-1]}" != --* ]]; then
                        has_branch=true
                        break
                    fi
                    ((j++))
                done
                
                if [[ "$has_branch" == false ]]; then
                    # Complete with git branches
                    COMPREPLY=($(compgen -W "$(_get_git_branches)" -- "$cur"))
                else
                    # Complete with options
                    COMPREPLY=($(compgen -W "$create_opts" -- "$cur"))
                fi
            fi
            ;;
        list)
            # List command only takes global options
            COMPREPLY=($(compgen -W "--worktree-dir --containerfile --runtime" -- "$cur"))
            ;;
        delete|terminal)
            # For delete and terminal commands, complete with environment names and options
            local env_opts="--worktree-dir --containerfile --runtime"
            if [[ "$cur" == -* ]]; then
                COMPREPLY=($(compgen -W "$env_opts" -- "$cur"))
            else
                # Check if we already have an environment name
                local has_env=false
                local j=$((i+1))
                while [[ $j -lt $COMP_CWORD ]]; do
                    if [[ "${COMP_WORDS[$j]}" != -* && "${COMP_WORDS[$j-1]}" != --* ]]; then
                        has_env=true
                        break
                    fi
                    ((j++))
                done
                
                if [[ "$has_env" == false ]]; then
                    # Complete with environment names
                    COMPREPLY=($(compgen -W "$(_get_environments)" -- "$cur"))
                else
                    # Complete with options
                    COMPREPLY=($(compgen -W "$env_opts" -- "$cur"))
                fi
            fi
            ;;
        *)
            # Default completion with commands and global options
            COMPREPLY=($(compgen -W "$commands $global_opts" -- "$cur"))
            ;;
    esac
}

# Register the completion function
complete -F _cc_buddy_completion cc-buddy

# Also register for common alternative names
complete -F _cc_buddy_completion ./cc-buddy