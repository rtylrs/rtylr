import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.portal.v1;

/**
 * One navigation area in display order. Hidden items keep their place so
 * showing them again puts them back where they were.
 */
model PortalArrangement {
  top: string[];

  /** Items pinned to the bottom of the app bar; empty for the sidebar. */
  bottom: string[];

  hidden: string[];

  /** False while the area follows the rtylr default. */
  customized: boolean;
}

model PortalLayout {
  rail: PortalArrangement;
  workspace: PortalArrangement;
}

model GetPortalLayoutRequest {}

model GetPortalLayoutResponse {
  layout: PortalLayout;
}

model SavePortalLayoutRequest {
  rail?: PortalArrangement;
  workspace?: PortalArrangement;
}

model SavePortalLayoutResponse {
  layout: PortalLayout;
}

model ResetPortalLayoutRequest {
  /** "rail", "workspace", or empty for both. */
  area: string;
}

model ResetPortalLayoutResponse {
  layout: PortalLayout;
}
