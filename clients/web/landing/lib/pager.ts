import { APP_URLS } from '@voxire/config';

/**
 * The pager's working brand name. Copy takes it through `{{brand}}`
 * interpolation, so renaming the product is this one line.
 */
export const PAGER_BRAND = 'Rennili';

/**
 * The instant try page in the Portal (public, no login). The route is
 * provisional until the self-serve signup work names it.
 */
export const PAGER_TRY_URL = `${APP_URLS.dash}/pager/try`;
