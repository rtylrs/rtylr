import React from 'react';
import MenuLayoutCore, { MenuLayoutCoreProps } from '../MenuLayoutCore';
import { MenuLayoutConfig } from '../../types';

type GridLayoutProps = Omit<MenuLayoutCoreProps, 'config'> & {
  config?: Partial<MenuLayoutConfig>;
};

const GRID_DEFAULTS: Partial<MenuLayoutConfig> = {
  items: {
    layout: 'grid',
    columns: 2,
    imagePosition: 'top',
    imageSize: 'md',
    showDescription: true,
    showBadges: true,
    cardRadius: 'md',
  },
};

const GridLayout: React.FC<GridLayoutProps> = ({ config, ...rest }) => (
  <MenuLayoutCore
    {...rest}
    config={{
      ...GRID_DEFAULTS,
      ...config,
      items: { ...GRID_DEFAULTS.items!, ...config?.items },
      header: config?.header,
      nav: config?.nav,
      pages: config?.pages,
    }}
  />
);

export default GridLayout;
