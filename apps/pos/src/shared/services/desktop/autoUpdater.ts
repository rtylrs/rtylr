import { check, type DownloadEvent } from '@tauri-apps/plugin-updater';
import { relaunch } from '@tauri-apps/plugin-process';
import { toast } from 'sonner';
import { isTauri } from '../../utils/platform';

const INITIAL_CHECK_DELAY_MS = 2_500;
// Terminals stay open for weeks - a single per-process check would never
// see updates. Re-check periodically.
const RECHECK_INTERVAL_MS = 4 * 60 * 60 * 1000;
// Progress chunks arrive hundreds of times for a large download; only
// refresh the toast on whole-percent steps.
const PROGRESS_TOAST_STEP_PCT = 1;

let hasStarted = false;

export const runDesktopAutoUpdater = (): void => {
  if (!isTauri() || hasStarted) return;

  hasStarted = true;
  window.setTimeout(() => {
    void installAvailableUpdate();
    window.setInterval(() => void installAvailableUpdate(), RECHECK_INTERVAL_MS);
  }, INITIAL_CHECK_DELAY_MS);
};

const installAvailableUpdate = async (): Promise<void> => {
  let toastId: string | number | undefined;
  let lastReportedPct = -PROGRESS_TOAST_STEP_PCT;

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
      if (pct - lastReportedPct < PROGRESS_TOAST_STEP_PCT) return;
      lastReportedPct = pct;
      toast.loading(`Installing rtylr POS ${update.version}... ${pct}%`, { id: toastId });
    });

    toast.success('Update installed. Restarting POS...', { id: toastId });
    await relaunch();
  } catch (error) {
    console.warn('[desktop-updater] update check failed', error);
    if (toastId) {
      toast.error('Unable to install POS update', { id: toastId });
    } else {
      // A failed check previously vanished silently; surface it so a
      // terminal stuck on an old version is diagnosable.
      toast.error('Update check failed - will retry in 4 hours');
    }
  }
};
