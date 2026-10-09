import { redirect } from '@tanstack/react-router';
import { canManageDash } from './canManageDash';
import { useUserStore } from '../stores/userStore';

export const requireDashManagement = () => {
  const user = useUserStore.getState().user;
  // Deny until Dash access has been hydrated; BusinessRoute handles the session.
  if (!canManageDash(user)) {
    throw redirect({ to: '/' });
  }
};
