import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.github.v2;

/** GitHubConnection is the business's GitHub App installation, if any. */
model GitHubConnection {
  /**
   * configured is false when the server has no GitHub App credentials, so
   * connecting is impossible until an operator sets them.
   */
  configured: boolean;

  /** connected is true when the business has stored an installation. */
  connected: boolean;

  /** account_login is the GitHub organisation or user the app is installed on. */
  account_login: string;

  /** account_type is "Organization" or "User", as GitHub reports it. */
  account_type: string;

  /**
   * status is the installation's health as GitHub reports it when this
   * response is built: "active", "suspended", "removed" (uninstalled on
   * GitHub), or "unknown" when GitHub could not be reached. Empty when not
   * connected.
   */
  status: string;

  /** connected_at is when the installation was stored. */
  connected_at: string;

  /** can_manage is true when the caller may connect or disconnect GitHub. */
  can_manage: boolean;

  /**
   * last_webhook_at is when the last signed webhook for this installation
   * arrived. Empty when none has arrived yet.
   */
  last_webhook_at: string;

  /** last_webhook_event is the X-GitHub-Event name of that webhook. */
  last_webhook_event: string;

  /**
   * linking_enabled is true when members can link their GitHub accounts:
   * the server has the app's OAuth client credentials.
   */
  linking_enabled: boolean;
}

/** GitHubAccountLink is the caller's own GitHub account in this business. */
model GitHubAccountLink {
  /** linked is false when the caller has not linked an account. */
  linked: boolean;

  /** github_user_id is GitHub's numeric user id. */
  @encode(string)
  github_user_id: int64;

  /** login is the GitHub username, without the \@. */
  login: string;

  /** avatar_url is the GitHub avatar, as GitHub reported it when linking. */
  avatar_url: string;

  /** linked_at is when the account was linked. */
  linked_at: string;
}

/** GitHubRepository is one repository the installation can access. */
model GitHubRepository {
  /** id is GitHub's numeric repository id. */
  @encode(string)
  id: int64;

  /** full_name is "owner/name". */
  full_name: string;

  /** private is true for a private repository. */
  private: boolean;
}

/** ProjectRepository is a repository a project has picked. */
model ProjectRepository {
  /** repository_id is GitHub's numeric repository id. */
  @encode(string)
  repository_id: int64;

  /** full_name is "owner/name" as stored when the repository was picked. */
  full_name: string;

  /** private is true for a private repository. */
  private: boolean;

  /** created_at is when the project picked the repository. */
  created_at: string;
}

model GetGitHubConnectionRequest {}

model GetGitHubConnectionResponse {
  connection: GitHubConnection;
}

model StartGitHubInstallRequest {}

model StartGitHubInstallResponse {
  /**
   * install_url is the GitHub page that installs the app. It carries a
   * signed state that expires after ten minutes.
   */
  install_url: string;
}

model CompleteGitHubInstallRequest {
  /** installation_id is the value GitHub appended to the Setup URL. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  installation_id?: int64;

  /** state is the value GitHub echoed back to the Setup URL. */
  @Onkt.required
  state: string;
}

model CompleteGitHubInstallResponse {
  connection: GitHubConnection;
}

model DisconnectGitHubRequest {}

model DisconnectGitHubResponse {
  /**
   * uninstalled is true when the app was also removed on GitHub. When false
   * the stored connection is gone but the app must be uninstalled on GitHub
   * by hand.
   */
  uninstalled: boolean;
}

model ListGitHubRepositoriesRequest {
  /**
   * project_id is the project the caller is picking repositories for. The
   * caller must be an admin of it.
   */
  @query("project_id")
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;
}

model ListGitHubRepositoriesResponse {
  repositories: GitHubRepository[];
}

model ListProjectRepositoriesRequest {
  @query("project_id")
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;
}

model ListProjectRepositoriesResponse {
  repositories: ProjectRepository[];
}

model SetProjectRepositoriesRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;

  /**
   * repository_ids replaces the project's whole selection. Every id must be
   * a repository the installation can access. Empty clears the selection.
   */
  @Onkt.encode("string")
  repository_ids: int64[];
}

model SetProjectRepositoriesResponse {
  repositories: ProjectRepository[];
}

model GetMyGitHubLinkRequest {}

model GetMyGitHubLinkResponse {
  link: GitHubAccountLink;
}

model StartGitHubLinkRequest {}

model StartGitHubLinkResponse {
  /**
   * authorize_url is the GitHub page that asks the member to authorize the
   * app. It carries a signed state that expires after ten minutes and is
   * bound to this business and member.
   */
  authorize_url: string;
}

model CompleteGitHubLinkRequest {
  /** code is the value GitHub appended to the callback URL. */
  @Onkt.required
  @minLength(1)
  @maxLength(256)
  code: string;

  /** state is the value GitHub echoed back to the callback URL. */
  @Onkt.required
  @minLength(1)
  @maxLength(512)
  state: string;
}

model CompleteGitHubLinkResponse {
  link: GitHubAccountLink;
}

model UnlinkGitHubRequest {}

model UnlinkGitHubResponse {}

model ListTaskPullRequestsRequest {
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;
}

/** GitHubCheck is one CI check on a pull request's head commit. */
model GitHubCheck {
  name: string;

  /** status is queued, in_progress or completed. */
  status: string;

  /** conclusion is set once completed: success, failure, neutral, skipped, cancelled, timed_out... */
  conclusion: string;

  /** details_url links to the check's log. */
  details_url: string;

  /** updated_at is RFC 3339. */
  updated_at: string;
}

/** TaskPullRequest is a pull request linked to a task. */
model TaskPullRequest {
  url: string;

  /** ref is owner/repo#number. */
  ref: string;

  title: string;

  /** state is open, draft, closed or merged. */
  state: string;

  /** known is false while no webhook has described the pull request yet. */
  known: boolean;

  /** checks_state is failing, running, passing or none. */
  checks_state: string;

  checks: GitHubCheck[];
}

model ListTaskPullRequestsResponse {
  pull_requests: TaskPullRequest[];
}
