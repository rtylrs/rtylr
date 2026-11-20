import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.github.v2;

/**
 * GitHubService connects a business to GitHub through a GitHub App and lets
 * a project pick repositories from that installation. Business-scoped: the
 * business is resolved from auth. Tokens and the app key never leave the
 * server.
 */
@route("/flow/github/v2")
interface GitHubService {
  /** Returns the business's installation and its health, read from GitHub. */
  @get
  @route("/connection")
  getGitHubConnection(
    ...GetGitHubConnectionRequest,
  ): GetGitHubConnectionResponse;

  /** Returns the GitHub URL that installs the app. Business admins only. */
  @post
  @route("/install")
  startGitHubInstall(...StartGitHubInstallRequest): StartGitHubInstallResponse;

  /**
   * Stores the installation GitHub redirected back with, after checking the
   * signed state and the installation itself. Business admins only.
   */
  @post
  @route("/install/complete")
  completeGitHubInstall(
    ...CompleteGitHubInstallRequest,
  ): CompleteGitHubInstallResponse;

  /**
   * Removes the installation and every project's repository selection.
   * Business admins only.
   */
  @delete
  @route("/connection")
  disconnectGitHub(...DisconnectGitHubRequest): DisconnectGitHubResponse;

  /** Lists the repositories the installation can access. Project admins only. */
  @get
  @route("/repositories")
  listGitHubRepositories(
    ...ListGitHubRepositoriesRequest,
  ): ListGitHubRepositoriesResponse;

  /** Lists the repositories a project has picked. */
  @get
  @route("/project-repositories")
  listProjectRepositories(
    ...ListProjectRepositoriesRequest,
  ): ListProjectRepositoriesResponse;

  /** Replaces a project's repository selection. Project admins only. */
  @put
  @route("/project-repositories")
  setProjectRepositories(
    ...SetProjectRepositoriesRequest,
  ): SetProjectRepositoriesResponse;

  /** Returns the caller's linked GitHub account in this business. */
  @get
  @route("/link")
  getMyGitHubLink(...GetMyGitHubLinkRequest): GetMyGitHubLinkResponse;

  /**
   * Returns the GitHub URL where the caller authorizes the app, so commits
   * and pull requests can be matched to them. Needs a connected business.
   */
  @post
  @route("/link")
  startGitHubLink(...StartGitHubLinkRequest): StartGitHubLinkResponse;

  /**
   * Exchanges the code GitHub redirected back with, after checking the
   * signed state, and stores the GitHub account id and login. The user
   * token is used once to read the account and is not stored.
   */
  @post
  @route("/link/complete")
  completeGitHubLink(...CompleteGitHubLinkRequest): CompleteGitHubLinkResponse;

  /** Removes the caller's linked GitHub account. */
  @delete
  @route("/link")
  unlinkGitHub(...UnlinkGitHubRequest): UnlinkGitHubResponse;
}
