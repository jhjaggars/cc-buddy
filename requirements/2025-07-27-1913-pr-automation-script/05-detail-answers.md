# Detail Answers

## Q1: Should the script store environment state in a hidden directory (.cc-buddy) in the repository root?
**Answer:** Yes

## Q2: Should the script detect container runtime (docker/podman) automatically or require explicit configuration?
**Answer:** Auto detect

## Q3: Should worktrees be created in a subdirectory like `.worktrees/` or alongside the main repository?
**Answer:** No, we should not default to putting worktrees inside the git repo. We should allow the user to specify where they want worktrees to go, and choose ~/.worktrees as a default. So there should be an argument for the path to worktrees, or we should also accept an ENV variable GIT_WORKTREES_DIR

## Q4: Should the script validate that the specified branch exists remotely before creating the worktree?
**Answer:** No, the branch doesn't need to be remote to make a worktree. In fact the pattern is likely to be make a new local branch and then make a worktree for it

## Q5: Should the script automatically pull the latest changes for the branch when creating the environment?
**Answer:** No, the user is responsible for making sure that the branch is up to date before starting the environment

## Q6: Should the container be started in detached mode or attached to the terminal for immediate use?
**Answer:** Detached

## Q7: Should the script provide a way to attach to an existing running environment?
**Answer:** No, the user can use existing podman or docker commands to attach to the container while it runs

## Q8: Should the script handle the case where containerfile.dev doesn't exist by offering to create one?
**Answer:** No, we should just error with a message saying that the user must create one or specify the path to the containerfile they want to use if it isn't Containerfile.dev

## Q9: Should environment names include the repository name to support multiple repositories?
**Answer:** Yes

## Q10: Should the script verify GitHub token validity before creating the environment?
**Answer:** No