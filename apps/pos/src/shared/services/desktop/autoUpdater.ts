import { check, type DownloadEvent } from '@tauri-apps/plugin-updater';
import { relaunch } from '@tauri-apps/plugin-process';
import { toast } from 'sonner';
import { isTauri } from '../../utils/platform';

let hasStarted = false;

export const runDesktopAutoUpdater = (): void => {
  if (!isTauri() || hasStarted) return;

  hasStarted = true;
  window.setTimeout(() => {
    void installAvailableUpdate();
  }, 2500);
};

const installAvailableUpdate = async (): Promise<void> => {
  let toastId: string | number | undefined;

  try {
    const update = await check({ timeout: 10_000 });
    if (!update) return;

    let downloaded = 0;
    let contentLength: number | undefined;

    toastId = toast.loading(`Installing rtylr POS ${update.version}...`);

    await update.downloadAndInstall((event: DownloadEvent) => {
      if (event.event === 'Started') {
        contentLength = event.data.contentLength;
        return;
      }

      if (event.event !== 'Progress') return;

      downloaded += event.data.chunkLength;
      if (!contentLength) return;

      const pct = Math.min(99, Math.round((downloaded / contentLength) * 100));
      toast.loading(`Installing rtylr POS ${update.version}... ${pct}%`, { id: toastId });
    });

    toast.success('Update installed. Restarting POS...', { id: toastId });
    await relaunch();
  } catch (error) {
    console.warn('[desktop-updater] update check failed', error);
    if (toastId) {
      toast.error('Unable to install POS update', { id: toastId });
    }
  }
};
