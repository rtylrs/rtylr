# Flow GitHub App: permissions and events

Flow's GitHub integration runs as one GitHub App per environment. The app's
settings live on GitHub (Settings > Developer settings > GitHub Apps), so a
change there is made by whoever administers the app, not by a deploy.

## Repository permissions

| Permission | Access | Used for |
| --- | --- | --- |
| Metadata | Read | Required by GitHub for every app |
| Pull requests | Read | Pull request titles, state and head commit (T-376) |
| Checks | Read | Check runs on pull request commits (T-376) |
| Commit statuses | Read | Statuses posted by CI systems that do not use checks (T-376) |

## Webhook events

Subscribe the app to:

- Installation and Installation repositories (connect, uninstall, suspend, repository access)
- Pull request
- Check run
- Status

GitHub only sends an event the app subscribes to and has the permission for.
After the permissions change, every organisation that installed the app is
asked to accept them; until it does, its deliveries do not include the new
events.

## What Flow does with them

`POST /flow/github/webhook` verifies `X-Hub-Signature-256` with
`GITHUB_APP_WEBHOOK_SECRET`, records the delivery id (a repeat is ignored) and
then:

- `pull_request` stores the PR's title, state (open, draft, closed, merged) and
  head commit in `flow_github_pull_requests`, and updates the state on task
  links to it.
- `check_run` and `status` store each check on a commit in `flow_github_checks`.
  A late "in progress" delivery never undoes a completed check, and a re-run
  (a new check run with the same name) replaces the earlier result.

A task's PR link (`flow_task_links`, kind `github_pr`, ref `owner/repo#123`)
joins to the stored pull request. Tasks show the worst state of their open
pull requests' checks: failing, running or passing.

Pull requests linked before the app received their events show "Waiting for
GitHub" until the next push or check on them.
