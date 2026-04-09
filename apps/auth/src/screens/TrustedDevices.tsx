import React from 'react';
import { motion } from 'framer-motion';
import { Monitor, Smartphone, Loader2, Trash2 } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import BackLink from '../components/BackLink';
import { useDevices, useRemoveDevice } from '../hooks/useDevices';
import notify from '../lib/notify/notify';

const TrustedDevices: React.FC = () => {
  const { data: devices, isLoading, refetch } = useDevices();
  const removeMutation = useRemoveDevice();

  const handleRemove = (id: string) => {
    removeMutation.mutate(
      { deviceId: id },
      {
        onSuccess: () => {
          notify.success('Device removed');
          refetch();
        },
      }
    );
  };

  return (
    <AuthShell
      title="Trusted devices"
      subtitle="Devices you've marked as trusted skip two-factor verification."
      footer={
        <div className="flex justify-center">
          <BackLink label="Back to account" />
        </div>
      }
    >
      <div className="flex flex-col gap-3">
        {isLoading && (
          <div className="flex items-center justify-center py-8 text-white/40">
            <Loader2 className="h-5 w-5 animate-spin" />
          </div>
        )}

        {devices && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="flex flex-col gap-2.5"
          >
            {devices.map((d) => {
              const Icon = d.os?.startsWith('iOS') ? Smartphone : Monitor;
              return (
                <div
                  key={d.id}
                  className="flex items-start gap-3.5 rounded-xl border border-white/10 bg-white/[0.03] p-4 hover:bg-white/[0.05] transition-colors"
                >
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-white/[0.05] border border-white/10">
                    <Icon className="h-4.5 w-4.5 text-white/75" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-semibold text-white truncate">
                      {d.name}
                    </p>
                    <p className="mt-0.5 text-xs text-white/50">
                      {d.browser} · {d.os}
                    </p>
                    <p className="mt-0.5 text-[11px] text-white/40">
                      Last used {d.lastUsed} · Trusted {d.trustedSince}
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={() => handleRemove(d.id)}
                    aria-label="Remove trusted device"
                    className="shrink-0 p-2 rounded-lg text-white/40 hover:text-red-300 hover:bg-red-500/10 transition-colors"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              );
            })}
            {devices.length === 0 && (
              <p className="text-center text-sm text-white/50 py-6">
                You don't have any trusted devices yet.
              </p>
            )}
          </motion.div>
        )}
      </div>
    </AuthShell>
  );
};

export default TrustedDevices;
