import { useLocation } from 'react-router-dom';
import { getLayoutTypeForRoute } from '../utils/layoutConfig';
import { LayoutType } from '../layouts/types';

export const useLayout = (): LayoutType => {
  const location = useLocation();
  return getLayoutTypeForRoute(location.pathname);
};

