import React from 'react';
import { motion } from 'framer-motion';
import { ShieldAlert, Loader2 } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import SessionCard from '../components/SessionCard';
import AuthButton from '../components/AuthButton';
import BackLink from '../components/BackLink';
import { useSessions, useRevokeSession, useRevokeAllSessions } from '../hooks/useSessions';
import notify from '../lib/notify/notify';

const Sessions: React.FC = () => {
  const { data: sessions, isLoading, refetch } = useSessions();
  const revokeOne = useRevokeSession();
  const revokeAll = useRevokeAllSessions();

  const handleRevoke = (id: string) => {
    revokeOne.mutate(
      { sessionId: id },
      {
        onSuccess: () => {
          notify.success('Session revoked');
          refetch();
        },
      }
    );
  };

  const handleRevokeAll = () => {
    revokeAll.mutate(undefined, {
      onSuccess: () => {
        notify.success('All other sessions revoked');
        refetch();
      },
    });
  };

  return (
    <AuthShell
      title="Active sessions"
      subtitle="Devices and locations where you're currently signed in."
      footer={
        <div className="flex justify-center">
          <BackLink label="Back to account" />
        </div>
      }
    >
      <div className="flex flex-col gap-4">
        {isLoading && (
          <div className="flex items-center justify-center py-8 text-white/40">
            <Loader2 className="h-5 w-5 animate-spin" />
          </div>
        )}

        {sessions && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="flex flex-col gap-2.5"
          >
            {sessions.map((s) => (
              <SessionCard key={s.id} session={s} onRevoke={handleRevoke} />
            ))}
          </motion.div>
        )}

        {sessions && sessions.length > 1 && (
          <div className="flex items-start gap-2.5 rounded-xl border border-amber-400/25 bg-amber-500/5 p-3.5 mt-2">
            <ShieldAlert className="h-4 w-4 text-amber-300 mt-0.5 shrink-0" />
            <div className="flex-1">
              <p className="text-xs text-amber-100 leading-relaxed">
                If you don't recognize a session, revoke it and change your
                password immediately.
              </p>
            </div>
          </div>
        )}

        <AuthButton
          variant="outline"
          onClick={handleRevokeAll}
          loading={revokeAll.isPending}
          disabled={!sessions || sessions.length <= 1}
        >
          Sign out of all other sessions
        </AuthButton>
      </div>
    </AuthShell>
  );
};

export default Sessions;
