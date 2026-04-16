Paddle Library Github Workflows
===============================

These workflows are defined in the `PaddleHQ/go-library-template` repository, which uses the `repo-file-sync-action` to 
synchronize the workflows into each library.

Any changes to workflows should be made in the `PaddleHQ/go-library-template` repository.

Private repos and access
------------------------

These actions are used in both public and private repositories, and as such need to manage secrets for pulling private
dependencies in our private repos.

The `setup-go` composite actions 

Dependabot
----------

GitHub has limitations in the way that secrets are available from PRs created via Dependabot. As such there are some
complexities in how we manage the source of workflows run in PRs.

For details see the 
 - [GitHub Actions: Workflows triggered by Dependabot PRs will run with read-only permissions](https://github.blog/changelog/2021-02-19-github-actions-workflows-triggered-by-dependabot-prs-will-run-with-read-only-permissions/)
 - [Keeping your GitHub Actions and workflows secure](https://securitylab.github.com/resources/github-actions-preventing-pwn-requests/)

You cannot use secrets in a `pull_request` workflow run by a dependabot PR. Instead, for dependabot PRs we use the 
`pull_request_target`, which runs the workflow in the context of the base branch, and not the PR branch.

As such there are four different cases to consider:

1. Pull request by user in normal library - we use branch workflows, this allows upgrading and changing of the 
   workflows.
2. Pull request by dependabot in normal library - we use the workflows defined in main - this is fine as long as we are 
   not changing workflows.
3. Pull request by user in template - use branch workflows so we can test changes.
4. Pull request by dependabot in template - use branch workflows so we can validate the updates to workflows, however as
   we have no private dependencies in the template then the key isn't needed.

Dependabot PRs for workflow actions should only be configured in the `PaddleHQ/go-template-library` repository. This
will allow the AppEx team to review the upgrades and changes to the workflows before they are merged. Once merged then
the repo-file-sync-action will update the workflows in the downstream libraries.
