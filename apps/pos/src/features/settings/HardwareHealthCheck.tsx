import React, { useState } from 'react';
import { checkHardwareHealth } from '../../shared/services/hardware/hardwareHealth';
import type { HardwareHealthReport, DeviceStatus } from '../../shared/services/hardware/hardwareHealth';
import type { PrinterSettings, KdsHardwareSettings } from '../../shared/stores/settingsStore';

interface Props {
  printer: PrinterSettings;
  kds: KdsHardwareSettings;
  showCashDrawer?: boolean;
  showKds?: boolean;
}

type CheckState = 'idle' | 'checking' | 'done' | 'error';

function StatusBadge({ status }: { status: DeviceStatus['status'] }) {
  if (status === 'ok')
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-green-500/15 px-2 py-0.5 text-xs font-semibold text-green-400">
        <span className="h-1.5 w-1.5 rounded-full bg-green-400" />
        OK
      </span>
    );
  if (status === 'unreachable')
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-red-500/15 px-2 py-0.5 text-xs font-semibold text-red-400">
        <span className="h-1.5 w-1.5 rounded-full bg-red-400" />
        Unreachable
      </span>
    );
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-zinc-700 px-2 py-0.5 text-xs font-semibold text-zinc-400">
      <span className="h-1.5 w-1.5 rounded-full bg-zinc-500" />
      Not configured
    </span>
  );
}

function DeviceRow({ label, device }: { label: string; device: DeviceStatus }) {
  return (
    <div className="flex items-start justify-between gap-4 rounded-lg bg-zinc-800/60 px-4 py-3">
      <div className="min-w-0 flex-1">
        <p className="text-sm font-medium text-white">{label}</p>
        <p className="mt-0.5 text-xs text-zinc-400">{device.message}</p>
      </div>
      <StatusBadge status={device.status} />
    </div>
  );
}

export function HardwareHealthCheck({
  printer,
  kds,
  showCashDrawer = true,
  showKds = true,
}: Props) {
  const [state, setState] = useState<CheckState>('idle');
  const [report, setReport] = useState<HardwareHealthReport | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function run() {
    setState('checking');
    setError(null);
    setReport(null);
    try {
      const result = await checkHardwareHealth(printer, kds);
      setReport(result);
      setState('done');
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
      setState('error');
    }
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-zinc-200">Hardware health</span>
        <button
          onClick={run}
          disabled={state === 'checking'}
          className="rounded-md bg-zinc-700 px-3 py-1.5 text-xs font-medium text-white transition hover:bg-zinc-600 disabled:cursor-wait disabled:opacity-50"
        >
          {state === 'checking' ? 'Checking…' : 'Test connections'}
        </button>
      </div>

      {state === 'done' && report && (
        <div className="space-y-2">
          <DeviceRow label="Printer" device={report.printer} />
          {showCashDrawer && <DeviceRow label="Cash drawer" device={report.drawer} />}
          {showKds && <DeviceRow label="KDS display" device={report.kds} />}
        </div>
      )}

      {state === 'error' && error && (
        <p className="rounded-lg bg-red-500/10 px-4 py-2 text-xs text-red-400">{error}</p>
      )}
    </div>
  );
}
