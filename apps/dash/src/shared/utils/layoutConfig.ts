import { LayoutType } from '../../layouts/types';

/**
 * Maps route paths to their layout types
 * Routes matching these patterns will use the specified layout
 */
export const getLayoutTypeForRoute = (pathname: string): LayoutType => {
  // Settings routes use custom sidebar layout
  if (pathname.startsWith('/settings')) {
    return 'custom-sidebar';
  }

  // Default layout for all other routes
  return 'default';
};
