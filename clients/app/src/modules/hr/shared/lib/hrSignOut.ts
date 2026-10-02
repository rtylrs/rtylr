import { signOut } from '@voxire/auth';
import { useHRConfigStore } from '../stores/hrConfigStore';

export function signOutFromHR(): void {
  useHRConfigStore.getState().clearConfig();
  signOut('hr');
}
