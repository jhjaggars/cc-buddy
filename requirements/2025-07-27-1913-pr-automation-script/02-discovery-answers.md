# Discovery Answers

## Q1: Will this script need to work with different container technologies (Docker, Podman, etc.)?
**Answer:** Yes - Docker and Podman should work

## Q2: Should the script automatically detect the GitHub repository URL from the current git remote?
**Answer:** Yes

## Q3: Will users need to specify which PR branch to work on, or should it auto-detect from GitHub?
**Answer:** Manual specification - for the first iteration just want to create a worktree and mount it into the container and make sure all the necessary tools can be used

## Q4: Should the development container use a pre-built image or build from a Dockerfile in the repo?
**Answer:** Build from a containerfile.dev or dockerfile.dev

## Q5: Will the script need to handle multiple GitHub tokens for different repositories?
**Answer:** No

## Q6: Should the script clean up Docker volumes when deleting an environment?
**Answer:** Yes - the worktree should be destroyed when the environment is destroyed

## Q7: Will users want to persist data between container restarts within the same environment?
**Answer:** Yes - the environment should remain available until the user asks the script to destroy it

## Q8: Should the script support custom Claude Code configuration per environment?
**Answer:** No - since the container will mount the worktree project specific customizations should be available in the dev container

## Q9: Will the script need to expose additional ports from the container for development servers?
**Answer:** Yes - ideally those can be expressed via the containerfile or dockerfile though

## Q10: Should environments be automatically named or allow custom naming?
**Answer:** Automatic naming