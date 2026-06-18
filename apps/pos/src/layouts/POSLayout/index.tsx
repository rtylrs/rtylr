import React from 'react';
import { CaretButton, CaretPanel, useCaretSettingsStore, useCaretStore } from '@voxire/caret';
import { Moon, Power, Sun } from 'lucide-react';
import { useThemeStore } from '../../shared/stores/themeStore';
import { useEmployeeSessionStore } from '../../shared/stores/employeeSessionStore';
import { useUserStore, businessDisplayName } from '../../shared/stores/userStore';
import { useSettingsStore } from '../../shared/stores/settingsStore';
import { RtylrWordmark } from '../../shared/components/RtylrWordmark';

interface POSLayoutProps {
  children: React.ReactNode;
}

const POSLayout: React.FC<POSLayoutProps> = ({ children }) => {
  const { theme, toggleTheme } = useThemeStore();
  const endSession = useEmployeeSessionStore((s) => s.endSession);
  const businessProfile = useUserStore((s) => s.businessProfile);
  const userIdValue = useUserStore((s) => s.userId);
  const configuredBusinessName = useSettingsStore((s) => s.business.name.trim());
  const businessName = configuredBusinessName || businessDisplayName(businessProfile);
  const caretOpen = useCaretStore((s) => s.isOpen);
  const caretWidth = useCaretSettingsStore((s) => s.panelWidth);
  const userId = Number(userIdValue ?? 0);
  const canUseCaret = Number.isFinite(userId) && userId > 0;

  return (
    <div className="flex flex-col h-screen bg-pos-bg text-pos-text" data-theme={theme}>
      <header className="flex items-center justify-between px-5 py-3 bg-pos-card border-b border-pos-border shrink-0">
        <div className="flex items-center gap-2.5">
          <RtylrWordmark className="text-xl" />
        </div>
        <div className="flex items-center gap-1">
          {businessName && <span className="text-xs text-pos-muted mr-2">{businessName}</span>}
          {canUseCaret && <CaretButton />}
          <button
            className="w-8 h-8 flex items-center justify-center rounded-lg text-pos-muted hover:bg-white/8 transition-colors"
            onClick={toggleTheme}
            aria-label="Toggle theme"
          >
            {theme === 'dark' ? (
              <Sun className="h-4 w-4" aria-hidden="true" />
            ) : (
              <Moon className="h-4 w-4" aria-hidden="true" />
            )}
          </button>
          <button
            className="w-8 h-8 flex items-center justify-center rounded-lg text-pos-muted hover:bg-white/8 transition-colors"
            onClick={endSession}
            aria-label="End session"
            title="End employee session"
          >
            <Power className="h-4 w-4" aria-hidden="true" />
          </button>
        </div>
      </header>
      <main className="flex flex-1 min-h-0 overflow-hidden">
        <div className="flex-1 min-w-0 overflow-hidden">{children}</div>
        {canUseCaret && (
          <div
            style={{
              width: caretOpen ? `${caretWidth}px` : '0px',
              flexShrink: 0,
              overflow: 'hidden',
              transition: caretOpen ? 'none' : 'width 0.25s cubic-bezier(0.4, 0, 0.2, 1)',
            }}
          >
            <CaretPanel
              variant="inline"
              appId="pos"
              userId={userId}
              apiBaseUrl={import.meta.env.VITE_API_URL ?? 'https://api.voxire.com/'}
              systemContext="User is in the POS app (terminal, orders, tables, receipts, inventory, customers, and staff workflows)."
            />
          </div>
        )}
      </main>
    </div>
  );
};

export default POSLayout;
