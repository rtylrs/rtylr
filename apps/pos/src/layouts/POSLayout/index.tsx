import React from 'react';
import { Moon, Power, Sun } from 'lucide-react';
import { useThemeStore } from '../../shared/stores/themeStore';
import { useEmployeeSessionStore } from '../../shared/stores/employeeSessionStore';
import { RtylrWordmark } from '../../shared/components/RtylrWordmark';

interface POSLayoutProps {
  children: React.ReactNode;
}

const POSLayout: React.FC<POSLayoutProps> = ({ children }) => {
  const { theme, toggleTheme } = useThemeStore();
  const endSession = useEmployeeSessionStore((s) => s.endSession);
  const employee = useEmployeeSessionStore((s) => s.employee);

  return (
    <div className="flex flex-col h-screen bg-pos-bg text-pos-text" data-theme={theme}>
      <header className="flex items-center justify-between px-5 py-3 bg-pos-card border-b border-pos-border shrink-0">
        <div className="flex items-center gap-2.5">
          <RtylrWordmark className="text-xl" />
        </div>
        <div className="flex items-center gap-1">
          {employee && <span className="text-xs text-pos-muted mr-2">{employee.name}</span>}
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
      <main className="flex-1 overflow-hidden">{children}</main>
    </div>
  );
};

export default POSLayout;
