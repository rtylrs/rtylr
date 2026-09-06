import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import {
  Search,
  Users,
  Clock,
  DollarSign,
  Building2,
  Calendar,
  ClipboardList,
  GraduationCap,
  Shield,
  FileText,
  Network,
  Settings as SettingsIcon,
  Rocket,
  LogOut,
  MessageSquare,
} from 'lucide-react';
import { useStaff } from '../../features/staff/useStaff';

interface Action {
  id: string;
  label: string;
  subtitle?: string;
  icon: React.ComponentType<{ size?: number }>;
  run: () => void;
  keywords?: string;
}

const useCommandPalette = () => {
  const [open, setOpen] = useState(false);
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const isMod = e.metaKey || e.ctrlKey;
      if (isMod && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setOpen((o) => !o);
      }
      if (e.key === 'Escape') setOpen(false);
    };
    const customOpen = () => setOpen(true);
    window.addEventListener('keydown', handler);
    window.addEventListener('hr:open-palette', customOpen);
    return () => {
      window.removeEventListener('keydown', handler);
      window.removeEventListener('hr:open-palette', customOpen);
    };
  }, []);
  return { open, setOpen };
};

const CommandPalette: React.FC = () => {
  const { open, setOpen } = useCommandPalette();
  const navigate = useNavigate();
  const { data: staff = [] } = useStaff();
  const [query, setQuery] = useState('');
  const [active, setActive] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open) {
      setQuery('');
      setActive(0);
      setTimeout(() => inputRef.current?.focus(), 10);
    }
  }, [open]);

  const baseActions: Action[] = useMemo(
    () => [
      {
        id: 'nav-dashboard',
        label: 'Dashboard',
        icon: Users,
        run: () => navigate({ to: '/' }),
        keywords: 'home',
      },
      { id: 'nav-staff', label: 'Staff', icon: Users, run: () => navigate({ to: '/staff' }) },
      { id: 'nav-shifts', label: 'Shifts', icon: Clock, run: () => navigate({ to: '/shifts' }) },
      {
        id: 'nav-schedule',
        label: 'Schedule',
        icon: Calendar,
        run: () => navigate({ to: '/schedule' }),
      },
      {
        id: 'nav-timesheets',
        label: 'Timesheets',
        icon: Clock,
        run: () => navigate({ to: '/timesheets' }),
      },
      {
        id: 'nav-payroll',
        label: 'Payroll',
        icon: DollarSign,
        run: () => navigate({ to: '/payroll' }),
      },
      {
        id: 'nav-departments',
        label: 'Departments',
        icon: Building2,
        run: () => navigate({ to: '/departments' }),
      },
      {
        id: 'nav-time-off',
        label: 'Time Off',
        icon: Calendar,
        run: () => navigate({ to: '/time-off' }),
      },
      {
        id: 'nav-new-time-off',
        label: 'New time off request',
        icon: Calendar,
        run: () => navigate({ to: '/time-off/new' }),
      },
      {
        id: 'nav-calendar',
        label: 'Time Off Calendar',
        icon: Calendar,
        run: () => navigate({ to: '/calendar' }),
      },
      {
        id: 'nav-reviews',
        label: 'Performance Reviews',
        icon: ClipboardList,
        run: () => navigate({ to: '/reviews' }),
      },
      {
        id: 'nav-one-on-ones',
        label: '1-on-1s',
        icon: MessageSquare,
        run: () => navigate({ to: '/one-on-ones' }),
      },
      {
        id: 'nav-onboarding',
        label: 'Onboarding',
        icon: Rocket,
        run: () => navigate({ to: '/onboarding' }),
      },
      {
        id: 'nav-offboarding',
        label: 'Offboarding',
        icon: LogOut,
        run: () => navigate({ to: '/offboarding' }),
      },
      {
        id: 'nav-documents',
        label: 'Documents',
        icon: FileText,
        run: () => navigate({ to: '/documents' }),
      },
      {
        id: 'nav-benefits',
        label: 'Benefits',
        icon: Shield,
        run: () => navigate({ to: '/benefits' }),
      },
      {
        id: 'nav-training',
        label: 'Training',
        icon: GraduationCap,
        run: () => navigate({ to: '/training' }),
      },
      {
        id: 'nav-org-chart',
        label: 'Org Chart',
        icon: Network,
        run: () => navigate({ to: '/org-chart' }),
      },
      {
        id: 'nav-settings',
        label: 'Settings',
        icon: SettingsIcon,
        run: () => navigate({ to: '/settings' }),
      },
    ],
    [navigate]
  );

  const staffActions: Action[] = useMemo(
    () =>
      staff.map((s) => ({
        id: `staff-${s.id}`,
        label: s.name,
        subtitle: `${s.role} • ${s.department}`,
        icon: Users,
        run: () => navigate({ to: '/staff/$id', params: { id: String(s.id) } }),
        keywords: s.email,
      })),
    [staff, navigate]
  );

  const all = useMemo(() => [...baseActions, ...staffActions], [baseActions, staffActions]);

  const filtered = useMemo(() => {
    if (!query) return all.slice(0, 8);
    const q = query.toLowerCase();
    return all
      .filter(
        (a) =>
          a.label.toLowerCase().includes(q) ||
          a.subtitle?.toLowerCase().includes(q) ||
          a.keywords?.toLowerCase().includes(q)
      )
      .slice(0, 10);
  }, [all, query]);

  const onRun = (a: Action) => {
    a.run();
    setOpen(false);
  };

  if (!open) return null;

  return (
    <div
      onClick={() => setOpen(false)}
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.6)',
        backdropFilter: 'blur(4px)',
        zIndex: 200,
        display: 'flex',
        justifyContent: 'center',
        paddingTop: 120,
      }}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          width: 560,
          maxHeight: '60vh',
          background: 'var(--hr-card)',
          border: '1px solid var(--hr-border)',
          borderRadius: 14,
          overflow: 'hidden',
          boxShadow: 'var(--hr-shadow)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            padding: '14px 18px',
            borderBottom: '1px solid var(--hr-border)',
          }}
        >
          <Search size={16} style={{ color: 'var(--hr-text-muted)' }} />
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setActive(0);
            }}
            onKeyDown={(e) => {
              if (e.key === 'ArrowDown') {
                e.preventDefault();
                setActive((a) => Math.min(a + 1, filtered.length - 1));
              } else if (e.key === 'ArrowUp') {
                e.preventDefault();
                setActive((a) => Math.max(a - 1, 0));
              } else if (e.key === 'Enter' && filtered[active]) {
                e.preventDefault();
                onRun(filtered[active]);
              }
            }}
            placeholder="Search people, pages, actions"
            style={{
              flex: 1,
              background: 'transparent',
              border: 'none',
              outline: 'none',
              color: 'var(--hr-text)',
              fontSize: 14,
            }}
          />
          <kbd
            style={{
              fontSize: 10,
              padding: '2px 6px',
              borderRadius: 4,
              border: '1px solid var(--hr-border)',
              color: 'var(--hr-text-muted)',
            }}
          >
            ESC
          </kbd>
        </div>
        <div style={{ overflowY: 'auto', maxHeight: 380 }}>
          {filtered.length === 0 && (
            <div
              style={{
                padding: 30,
                textAlign: 'center',
                color: 'var(--hr-text-muted)',
                fontSize: 13,
              }}
            >
              No matches
            </div>
          )}
          {filtered.map((a, i) => {
            const Icon = a.icon;
            const isActive = i === active;
            return (
              <button
                type="button"
                key={a.id}
                onClick={() => onRun(a)}
                onMouseEnter={() => setActive(i)}
                style={{
                  width: '100%',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 12,
                  padding: '10px 18px',
                  background: isActive ? 'var(--hr-hover)' : 'transparent',
                  border: 'none',
                  color: 'var(--hr-text)',
                  cursor: 'pointer',
                  textAlign: 'left',
                }}
              >
                <Icon size={14} />
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: 13, fontWeight: 600 }}>{a.label}</div>
                  {a.subtitle && (
                    <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>{a.subtitle}</div>
                  )}
                </div>
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
};

export default CommandPalette;
