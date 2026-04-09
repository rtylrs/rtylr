import React from 'react';
import { Monitor, Smartphone, Tablet, MapPin, Clock } from 'lucide-react';

export interface SessionInfo {
  id: string;
  device: string;
  browser?: string;
  os?: string;
  location?: string;
  ip?: string;
  lastActive: string;
  current?: boolean;
  type?: 'desktop' | 'mobile' | 'tablet';
}

export interface SessionCardProps {
  session: SessionInfo;
  onRevoke?: (id: string) => void;
}

const ICONS = {
  desktop: Monitor,
  mobile: Smartphone,
  tablet: Tablet,
};

const SessionCard: React.FC<SessionCardProps> = ({ session, onRevoke }) => {
  const Icon = ICONS[session.type ?? 'desktop'];

  return (
    <div className="flex items-start gap-3.5 rounded-xl border border-white/10 bg-white/[0.03] p-4 hover:bg-white/[0.05] transition-colors">
      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-white/[0.05] border border-white/10">
        <Icon className="h-4.5 w-4.5 text-white/75" />
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <p className="text-sm font-semibold text-white truncate">
            {session.device}
          </p>
          {session.current && (
            <span className="shrink-0 inline-flex items-center rounded-full bg-[#64C5B9]/15 px-2 py-0.5 text-[10px] font-semibold text-[#64C5B9] border border-[#64C5B9]/30">
              This device
            </span>
          )}
        </div>
        {(session.browser || session.os) && (
          <p className="mt-0.5 text-xs text-white/50 truncate">
            {[session.browser, session.os].filter(Boolean).join(' · ')}
          </p>
        )}
        <div className="mt-1.5 flex flex-wrap items-center gap-3 text-[11px] text-white/40">
          {session.location && (
            <span className="inline-flex items-center gap-1">
              <MapPin className="h-3 w-3" />
              {session.location}
              {session.ip && <span className="text-white/30">· {session.ip}</span>}
            </span>
          )}
          <span className="inline-flex items-center gap-1">
            <Clock className="h-3 w-3" />
            {session.lastActive}
          </span>
        </div>
      </div>
      {!session.current && onRevoke && (
        <button
          type="button"
          onClick={() => onRevoke(session.id)}
          className="shrink-0 text-[11px] font-semibold text-red-300 hover:text-red-200 hover:underline underline-offset-2"
        >
          Revoke
        </button>
      )}
    </div>
  );
};

export default SessionCard;
