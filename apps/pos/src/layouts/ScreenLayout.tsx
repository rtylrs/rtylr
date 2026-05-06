import React from 'react';
import { useNavigate } from 'react-router-dom';
import POSLayout from './POSLayout';
import QuickAccessSidebar from '../shared/components/QuickAccessSidebar';
import OfflineBadge from '../shared/components/OfflineBadge';

interface ScreenLayoutProps {
  title: string;
  subtitle?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
  showBack?: boolean;
}

const ScreenLayout: React.FC<ScreenLayoutProps> = ({
  title,
  subtitle,
  actions,
  children,
  showBack = true,
}) => {
  const navigate = useNavigate();
  return (
    <POSLayout>
      <div className="flex h-full overflow-hidden">
        <QuickAccessSidebar />
        <div className="flex-1 flex flex-col overflow-hidden">
          <div className="flex items-center justify-between gap-3 px-5 py-4 border-b border-pos-border shrink-0">
            <div className="flex items-center gap-3 min-w-0">
              {showBack && (
                <button
                  onClick={() => navigate(-1)}
                  className="w-10 h-10 rounded-lg bg-white/8 hover:bg-white/14 text-pos-text transition"
                  aria-label="Go back"
                >
                  ←
                </button>
              )}
              <div className="min-w-0">
                <h1 className="text-lg font-bold text-pos-text truncate">{title}</h1>
                {subtitle && <p className="text-xs text-pos-muted truncate">{subtitle}</p>}
              </div>
            </div>
            <div className="flex items-center gap-2">
              <OfflineBadge />
              {actions}
            </div>
          </div>
          <div className="flex-1 overflow-y-auto">{children}</div>
        </div>
      </div>
    </POSLayout>
  );
};

export default ScreenLayout;
