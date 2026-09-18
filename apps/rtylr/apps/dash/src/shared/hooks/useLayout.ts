import { useRouterState } from '@tanstack/react-router';
import { getLayoutTypeForRoute } from '../utils/layoutConfig';
import { LayoutType } from '../../layouts/types';

export const useLayout = (): LayoutType => {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  return getLayoutTypeForRoute(pathname);
};
