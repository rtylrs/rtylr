import React from 'react';
import MenuLayoutCore, { MenuLayoutCoreProps } from '../MenuLayoutCore';
import { MenuLayoutConfig } from '../../types';

type ListLayoutProps = Omit<MenuLayoutCoreProps, 'config'> & {
  config?: Partial<MenuLayoutConfig>;
};

const LIST_DEFAULTS: Partial<MenuLayoutConfig> = {
  items: {
    layout: 'list',
    columns: 2,
    imagePosition: 'right',
    imageSize: 'md',
    showDescription: true,
    showBadges: true,
    cardRadius: 'md',
  },
};

const ListLayout: React.FC<ListLayoutProps> = ({ config, ...rest }) => (
  <MenuLayoutCore
    {...rest}
    config={{
      ...LIST_DEFAULTS,
      ...config,
      items: { ...LIST_DEFAULTS.items!, ...config?.items },
      header: config?.header,
      nav: config?.nav,
      pages: config?.pages,
    }}
  />
);

export default ListLayout;
