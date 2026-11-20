import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.app.v1;

model App {
  id: string;
  name: string;
  description: string;
  icon: string;
  is_enabled: boolean;
}

model Addon {
  id: string;
  name: string;
  description: string;
  icon: string;
  price: float64;
  is_free: boolean;
  is_enabled: boolean;
  enabled_at?: string;
  disabled_at?: string;
}

model ListAppsRequest {}

model ListAppsResponse {
  apps: App[];
}

model ListAddonsRequest {}

model ListAddonsResponse {
  addons: Addon[];
}

model AddonStatusRequest {
  @Onkt.required
  addon_id: string;
}

model AddonStatusResponse {
  is_enabled: boolean;
}

model EnableAddonRequest {
  @Onkt.required
  addon_id: string;
}

model EnableAddonResponse {
  message: string;
}

model DisableAddonRequest {
  @Onkt.required
  addon_id: string;
}

model DisableAddonResponse {
  message: string;
}
