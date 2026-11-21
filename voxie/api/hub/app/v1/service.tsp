import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.app.v1;

/** AppService defines hub app and addon endpoints. */
@route("/hub/app/v1")
interface AppService {
  /** Returns installed apps for the active business. */
  @get
  @route("/apps")
  listApps(...ListAppsRequest): ListAppsResponse;

  /** Returns available addons. */
  @get
  @route("/addons")
  listAddons(...ListAddonsRequest): ListAddonsResponse;

  /** Returns the enabled state of an addon. */
  @get
  @route("/addons/{addon_id}/status")
  addonStatus(...AddonStatusRequest): AddonStatusResponse;

  /** Enables an addon for the active business. */
  @patch
  @route("/addons/{addon_id}/enable")
  enableAddon(...EnableAddonRequest): EnableAddonResponse;

  /** Disables an addon for the active business. */
  @patch
  @route("/addons/{addon_id}/disable")
  disableAddon(...DisableAddonRequest): DisableAddonResponse;
}
