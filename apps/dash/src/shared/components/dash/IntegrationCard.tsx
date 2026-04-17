import React from 'react';
import { Check } from 'lucide-react';
import type { Integration } from '../../hooks/queries/useIntegrations';
import { Button } from '../ui';

export interface IntegrationCardProps {
  integration: Integration;
  onConnect?: (id: string) => void;
  onDisconnect?: (id: string) => void;
  isLoading?: boolean;
}

export const IntegrationCard: React.FC<IntegrationCardProps> = ({
  integration,
  onConnect,
  onDisconnect,
  isLoading,
}) => {
  const isConnected = integration.status === 'connected';
  const isComingSoon = integration.status === 'coming_soon';

  return (
    <div
      className="rounded-xl p-4 flex flex-col gap-3"
      style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
    >
      <div className="flex items-start justify-between">
        <div
          className="h-10 w-10 rounded-lg flex items-center justify-center shrink-0"
          style={{
            background: 'var(--dash-hover)',
            border: '1px solid var(--dash-border)',
          }}
        >
          <span className="text-sm font-bold" style={{ color: 'var(--dash-text)' }}>
            {integration.logo}
          </span>
        </div>
        {isConnected && (
          <span
            className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase"
            style={{ background: 'rgba(16,185,129,0.15)', color: '#10b981' }}
          >
            <Check size={10} /> Connected
          </span>
        )}
        {isComingSoon && (
          <span
            className="px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase"
            style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-dim)' }}
          >
            Soon
          </span>
        )}
      </div>
      <div>
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          {integration.name}
        </p>
        <p
          className="text-xs mt-0.5 leading-relaxed line-clamp-2"
          style={{ color: 'var(--dash-text-muted)' }}
        >
          {integration.description}
        </p>
      </div>
      <div className="pt-1">
        {isConnected ? (
          <Button
            variant="secondary"
            size="sm"
            className="w-full"
            onClick={() => onDisconnect?.(integration.id)}
            disabled={isLoading}
          >
            Disconnect
          </Button>
        ) : isComingSoon ? (
          <Button variant="secondary" size="sm" className="w-full" disabled>
            Coming soon
          </Button>
        ) : (
          <Button
            variant="default"
            size="sm"
            className="w-full"
            onClick={() => onConnect?.(integration.id)}
            disabled={isLoading}
          >
            Connect
          </Button>
        )}
      </div>
    </div>
  );
};

export default IntegrationCard;
