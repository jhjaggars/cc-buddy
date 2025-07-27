# Expert Detail Questions

## Q1: Should the script store environment state in a hidden directory (.cc-buddy) in the repository root?
**Default if unknown:** Yes (follows standard tool conventions like .git, .vscode)

## Q2: Should the script detect container runtime (docker/podman) automatically or require explicit configuration?
**Default if unknown:** Auto-detect (check for podman first, fall back to docker)

## Q3: Should worktrees be created in a subdirectory like `.worktrees/` or alongside the main repository?
**Default if unknown:** Subdirectory (keeps workspace organized and separate from main repo)

## Q4: Should the script validate that the specified branch exists remotely before creating the worktree?
**Default if unknown:** Yes (prevents creation of environments for non-existent branches)

## Q5: Should the script automatically pull the latest changes for the branch when creating the environment?
**Default if unknown:** Yes (ensures development starts with current code)

## Q6: Should the container be started in detached mode or attached to the terminal for immediate use?
**Default if unknown:** Detached mode (allows script to complete and user to attach later)

## Q7: Should the script provide a way to attach to an existing running environment?
**Default if unknown:** Yes (essential for resuming work in existing environments)

## Q8: Should the script handle the case where containerfile.dev doesn't exist by offering to create one?
**Default if unknown:** No (fail with helpful error message pointing to documentation)

## Q9: Should environment names include the repository name to support multiple repositories?
**Default if unknown:** Yes (prevents conflicts when using script across multiple projects)

## Q10: Should the script verify GitHub token validity before creating the environment?
**Default if unknown:** No (delegate token validation to tools that actually use it)