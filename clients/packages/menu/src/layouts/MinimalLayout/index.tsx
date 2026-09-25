import React from 'react';
import MenuLayoutCore, { MenuLayoutCoreProps } from '../MenuLayoutCore';
import { MenuLayoutConfig } from '../../types';

type MinimalLayoutProps = Omit<MenuLayoutCoreProps, 'config'> & {
  config?: Partial<MenuLayoutConfig>;
};

const MINIMAL_DEFAULTS: Partial<MenuLayoutConfig> = {
  items: {
    layout: 'minimal',
    columns: 2,
    imagePosition: 'none',
    imageSize: 'sm',
    showDescription: true,
    showBadges: false,
    cardRadius: 'none',
  },
};

const MinimalLayout: React.FC<MinimalLayoutProps> = ({ config, enableCart = false, ...rest }) => (
  <MenuLayoutCore
    {...rest}
    enableCart={enableCart}
    config={{
      ...MINIMAL_DEFAULTS,
      ...config,
      items: { ...MINIMAL_DEFAULTS.items!, ...config?.items },
      header: config?.header,
      nav: config?.nav,
      pages: config?.pages,
    }}
  />
);

export default MinimalLayout;
