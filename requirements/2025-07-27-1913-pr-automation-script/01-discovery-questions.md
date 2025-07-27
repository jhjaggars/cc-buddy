# Discovery Questions

## Q1: Will this script need to work with different container technologies (Docker, Podman, etc.)?
**Default if unknown:** Yes (flexibility across different container runtimes is important for wide adoption)

## Q2: Should the script automatically detect the GitHub repository URL from the current git remote?
**Default if unknown:** Yes (reduces manual input and follows git conventions)

## Q3: Will users need to specify which PR branch to work on, or should it auto-detect from GitHub?
**Default if unknown:** Manual specification (gives users explicit control over which PR to work with)

## Q4: Should the development container use a pre-built image or build from a Dockerfile in the repo?
**Default if unknown:** Build from Dockerfile (allows repo-specific customization and dependencies)

## Q5: Will the script need to handle multiple GitHub tokens for different repositories?
**Default if unknown:** No (single token approach is simpler for initial implementation)

## Q6: Should the script clean up Docker volumes when deleting an environment?
**Default if unknown:** Yes (prevents disk space accumulation and maintains clean state)

## Q7: Will users want to persist data between container restarts within the same environment?
**Default if unknown:** Yes (preserves work and installed dependencies during development)

## Q8: Should the script support custom Claude Code configuration per environment?
**Default if unknown:** No (use default Claude Code settings for simplicity)

## Q9: Will the script need to expose additional ports from the container for development servers?
**Default if unknown:** Yes (common need for web development and testing)

## Q10: Should environments be automatically named or allow custom naming?
**Default if unknown:** Automatic naming (based on PR number/branch name for consistency)