import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using pos.menu_item.v1;

namespace hub.menu.v1;

model ListMenuRequest {}

model ListMenuResponse {
  items: MenuItem[];
}

model UpdateMenuRequest {
  @minItems(1)
  items: MenuItem[];
}

model UpdateMenuResponse {
  success: boolean;
}

model MenuBanner {
  @encode(string)
  id: int64;

  title: string;
  body: string;
  image_url: string;
  link_url: string;
  position: int32;
  is_active?: boolean;
  starts_at: string;
  ends_at: string;
  created_at: string;
  updated_at: string;
}

model ListMenuBannersRequest {}

model ListMenuBannersResponse {
  banners: MenuBanner[];
}

model CreateMenuBannerRequest {
  @Onkt.required
  title: string;

  body: string;
  image_url: string;
  link_url: string;
  position: int32;
  is_active?: boolean;
  starts_at: string;
  ends_at: string;
}

model CreateMenuBannerResponse {
  banner: MenuBanner;
}

model UpdateMenuBannerRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @Onkt.required
  title: string;

  body: string;
  image_url: string;
  link_url: string;
  position: int32;
  is_active: boolean;
  starts_at: string;
  ends_at: string;
}

model UpdateMenuBannerResponse {
  banner: MenuBanner;
}

model DeleteMenuBannerRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteMenuBannerResponse {
  success: boolean;
}
