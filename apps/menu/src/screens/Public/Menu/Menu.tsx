import React from 'react';
import Public from '../../../layouts/Public/public';
import RealMenu from '../../../features/menu/Menu/Menu';

/**
 * Wire the legacy `/:id` route (mongo-id-keyed QR codes from before
 * the POS migration) to the real menu implementation that lives in
 * `features/menu/Menu/Menu.tsx`. The screen layer used to render
 * `<Public>{null}</Public>`, which silently dropped the legacy URLs.
 *
 * New businesses go through `/b/:businessId` (PosMenu); this path
 * exists only so existing QR codes in the wild keep resolving.
 */
const Menu: React.FC = (): React.JSX.Element => {
  return (
    <Public>
      <RealMenu />
    </Public>
  );
};

export default Menu;
