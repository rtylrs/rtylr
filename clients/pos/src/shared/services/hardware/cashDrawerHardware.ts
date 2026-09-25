import { invoke } from '@tauri-apps/api/core';
import { canUseNativeHardware } from '../../utils/platform';

type InvokeCommand = (command: string, args?: Record<string, unknown>) => Promise<unknown>;

type DrawerResult = { status: 'opened' } | { status: 'skipped' } | { status: 'failed'; error: string };

interface OpenConfiguredCashDrawerParams {
  enabled: boolean;
  printerName: string;
  invokeCommand?: InvokeCommand;
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

export async function openConfiguredCashDrawer({
  enabled,
  printerName,
  invokeCommand = invoke,
}: OpenConfiguredCashDrawerParams): Promise<DrawerResult> {
  const trimmedPrinterName = printerName.trim();
  if (!enabled || !trimmedPrinterName || !canUseNativeHardware()) {
    return { status: 'skipped' };
  }

  try {
    await invokeCommand('open_cash_drawer', { printerName: trimmedPrinterName });
    return { status: 'opened' };
  } catch (error) {
    return { status: 'failed', error: errorMessage(error) };
  }
}
