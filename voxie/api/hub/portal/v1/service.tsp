import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.portal.v1;

/**
 * The signed-in person's arrangement of the Portal navigation: the order of
 * the app bar and the workspace sidebar, and what they hid. Caret changes it
 * through the same rules, so both stay in step.
 */
@route("/")
interface PortalLayoutService {
  @get
  @route("/portal/layout")
  getPortalLayout(...GetPortalLayoutRequest): GetPortalLayoutResponse;

  /** Saves the arrangement of the areas sent; areas left out keep theirs. */
  @put
  @route("/portal/layout")
  savePortalLayout(...SavePortalLayoutRequest): SavePortalLayoutResponse;

  /**
   * Restores the default arrangement of one area, or of every area when
   * none is named.
   */
  @post
  @route("/portal/layout/reset")
  resetPortalLayout(...ResetPortalLayoutRequest): ResetPortalLayoutResponse;
}
