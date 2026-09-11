/**
 * Shared UI primitives used across HR screens.
 * Keeps visual consistency with the existing Tailwind + CSS variable design.
 */
import React from 'react';
import { useTranslation } from 'react-i18next';

export const ACCENT = 'var(--hr-accent)';

export const QueryError: React.FC<{ onRetry: () => void; isRetrying?: boolean }> = ({
  onRetry,
  isRetrying = false,
}) => {
  const { t } = useTranslation();
  return (
    <div
      role="alert"
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 12,
        padding: 32,
      }}
    >
      <p style={{ margin: 0, color: 'var(--hr-text)' }}>{t('common.loadError')}</p>
      <Button variant="primary" onClick={onRetry} disabled={isRetrying}>
        {t('common.retry')}
      </Button>
    </div>
  );
};

export function useMediaQuery(query: string): boolean {
  const getMatch = () => (typeof window !== 'undefined' ? window.matchMedia(query).matches : false);
  const [matches, setMatches] = React.useState<boolean>(getMatch);

  React.useEffect(() => {
    if (typeof window === 'undefined') return undefined;
    const media = window.matchMedia(query);
    const listener = () => setMatches(media.matches);
    listener();
    if (typeof media.addEventListener === 'function') {
      media.addEventListener('change', listener);
      return () => media.removeEventListener('change', listener);
    }
    media.addListener(listener);
    return () => media.removeListener(listener);
  }, [query]);

  return matches;
}

export function useIsMobile(maxWidth = 768): boolean {
  return useMediaQuery(`(max-width: ${maxWidth}px)`);
}

/* ── Card ── */
export const Card: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  style,
  children,
  ...rest
}) => (
  <div
    style={{
      background: 'var(--hr-card)',
      border: '1px solid var(--hr-border)',
      borderRadius: 12,
      padding: 16,
      ...style,
    }}
    {...rest}
  >
    {children}
  </div>
);

/* ── Section header ── */
export const SectionHeader: React.FC<{
  title: string;
  subtitle?: string;
  right?: React.ReactNode;
}> = ({ title, subtitle, right }) => {
  const stacked = useIsMobile(720);

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'flex-start',
        justifyContent: 'space-between',
        flexWrap: 'wrap',
        gap: 12,
        marginBottom: 4,
      }}
    >
      <div>
        <h2 style={{ margin: 0, fontSize: 18, fontWeight: 700, color: 'var(--hr-text)' }}>
          {title}
        </h2>
        {subtitle && (
          <p style={{ margin: '4px 0 0', fontSize: 13, color: 'var(--hr-text-muted)' }}>
            {subtitle}
          </p>
        )}
      </div>
      {right && (
        <div
          style={{
            display: 'flex',
            flexWrap: 'wrap',
            gap: 8,
            width: stacked ? '100%' : 'auto',
            justifyContent: stacked ? 'flex-start' : 'flex-end',
          }}
        >
          {right}
        </div>
      )}
    </div>
  );
};

/* ── Button ── */
type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';
export const Button: React.FC<
  React.ButtonHTMLAttributes<HTMLButtonElement> & {
    variant?: ButtonVariant;
    icon?: React.ReactNode;
  }
> = ({ variant = 'secondary', icon, children, style, ...rest }) => {
  const base: React.CSSProperties = {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 6,
    minHeight: 36,
    padding: '8px 12px',
    borderRadius: 8,
    fontSize: 12,
    fontWeight: 600,
    cursor: 'pointer',
    transition: 'background 0.15s, color 0.15s, border-color 0.15s',
    border: '1px solid var(--hr-border)',
    background: 'transparent',
    color: 'var(--hr-text)',
    whiteSpace: 'nowrap',
  };
  if (variant === 'primary') {
    base.background = ACCENT;
    base.color = 'var(--hr-on-accent)';
    base.border = `1px solid ${ACCENT}`;
  } else if (variant === 'danger') {
    base.color = 'var(--hr-error)';
    base.border = '1px solid color-mix(in srgb, var(--hr-error) 35%, transparent)';
  } else if (variant === 'ghost') {
    base.border = '1px solid transparent';
    base.color = 'var(--hr-text-muted)';
  }
  return (
    <button type="button" style={{ ...base, ...style }} {...rest}>
      {icon}
      {children}
    </button>
  );
};

/* ── Input ── */
export const Input: React.FC<React.InputHTMLAttributes<HTMLInputElement>> = ({
  style,
  ...rest
}) => (
  <input
    style={{
      width: '100%',
      padding: '8px 10px',
      borderRadius: 8,
      border: '1px solid var(--hr-border)',
      background: 'var(--hr-input-bg)',
      color: 'var(--hr-text)',
      fontSize: 13,
      outline: 'none',
      ...style,
    }}
    {...rest}
  />
);

export const Textarea: React.FC<React.TextareaHTMLAttributes<HTMLTextAreaElement>> = ({
  style,
  ...rest
}) => (
  <textarea
    style={{
      width: '100%',
      padding: '8px 10px',
      borderRadius: 8,
      border: '1px solid var(--hr-border)',
      background: 'var(--hr-input-bg)',
      color: 'var(--hr-text)',
      fontSize: 13,
      outline: 'none',
      resize: 'vertical',
      minHeight: 80,
      fontFamily: 'inherit',
      ...style,
    }}
    {...rest}
  />
);

export const Select: React.FC<React.SelectHTMLAttributes<HTMLSelectElement>> = ({
  style,
  children,
  ...rest
}) => (
  <select
    style={{
      width: '100%',
      padding: '8px 10px',
      borderRadius: 8,
      border: '1px solid var(--hr-border)',
      background: 'var(--hr-input-bg)',
      color: 'var(--hr-text)',
      fontSize: 13,
      outline: 'none',
      ...style,
    }}
    {...rest}
  >
    {children}
  </select>
);

/* ── Label ── */
export const Label: React.FC<React.LabelHTMLAttributes<HTMLLabelElement>> = ({
  children,
  style,
  ...rest
}) => (
  <label
    style={{
      display: 'block',
      fontSize: 11,
      fontWeight: 600,
      color: 'var(--hr-text-muted)',
      textTransform: 'uppercase',
      letterSpacing: '0.05em',
      marginBottom: 8,
      ...style,
    }}
    {...rest}
  >
    {children}
  </label>
);

/* ── Pill / Badge ── */
export const Pill: React.FC<{
  color?: string;
  children: React.ReactNode;
  style?: React.CSSProperties;
}> = ({ color = ACCENT, children, style }) => (
  <span
    style={{
      display: 'inline-flex',
      alignItems: 'center',
      padding: '3px 8px',
      borderRadius: 999,
      fontSize: 10,
      fontWeight: 600,
      background: `${color}22`,
      color,
      border: `1px solid ${color}33`,
      ...style,
    }}
  >
    {children}
  </span>
);

/* ── Skeleton loader ── */
export const Skeleton: React.FC<{
  height?: number | string;
  width?: number | string;
  style?: React.CSSProperties;
}> = ({ height = 16, width = '100%', style }) => (
  <div
    style={{
      height,
      width,
      borderRadius: 8,
      background:
        'linear-gradient(90deg, var(--hr-hover) 25%, var(--hr-active) 50%, var(--hr-hover) 75%)',
      backgroundSize: '200% 100%',
      animation: 'hr-shimmer 1.5s ease-in-out infinite',
      ...style,
    }}
  />
);

/* ── Modal ── */
export const Modal: React.FC<{
  open: boolean;
  onClose: () => void;
  title?: string;
  children?: React.ReactNode;
  width?: number;
}> = ({ open, onClose, title, children, width = 480 }) => {
  if (!open) return null;
  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.55)',
        backdropFilter: 'blur(3px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 12,
        zIndex: 100,
      }}
      onClick={onClose}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          width: `min(calc(100vw - 24px), ${width}px)`,
          maxHeight: '85vh',
          overflowY: 'auto',
          background: 'var(--hr-card)',
          border: '1px solid var(--hr-border)',
          borderRadius: 14,
          padding: 20,
          boxShadow: 'var(--hr-shadow)',
        }}
      >
        {title && (
          <h3
            style={{ margin: '0 0 14px', fontSize: 15, fontWeight: 700, color: 'var(--hr-text)' }}
          >
            {title}
          </h3>
        )}
        {children}
      </div>
    </div>
  );
};

/* ── Confirm dialog ── */
export const ConfirmDialog: React.FC<{
  open: boolean;
  title: string;
  description?: string;
  confirmLabel?: string;
  variant?: 'danger' | 'primary';
  onConfirm: () => void;
  onCancel: () => void;
}> = ({
  open,
  title,
  description,
  confirmLabel = 'Confirm',
  variant = 'primary',
  onConfirm,
  onCancel,
}) => {
  const { t } = useTranslation();
  return (
    <Modal open={open} onClose={onCancel} title={title} width={380}>
      {description && (
        <p style={{ margin: '0 0 16px', fontSize: 13, color: 'var(--hr-text-muted)' }}>
          {description}
        </p>
      )}
      <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
        <Button variant="ghost" onClick={onCancel}>
          {t('common.cancel')}
        </Button>
        <Button variant={variant === 'danger' ? 'danger' : 'primary'} onClick={onConfirm}>
          {confirmLabel}
        </Button>
      </div>
    </Modal>
  );
};

/* ── Breadcrumbs ── */
export const Breadcrumbs: React.FC<{ items: Array<{ label: string; href?: string }> }> = ({
  items,
}) => (
  <nav
    style={{
      display: 'flex',
      alignItems: 'center',
      gap: 6,
      flexWrap: 'wrap',
      fontSize: 11,
      color: 'var(--hr-text-muted)',
      marginBottom: 12,
    }}
  >
    {items.map((item, i) => {
      const last = i === items.length - 1;
      return (
        <React.Fragment key={i}>
          {item.href && !last ? (
            <a href={item.href} style={{ color: 'var(--hr-text-muted)', textDecoration: 'none' }}>
              {item.label}
            </a>
          ) : (
            <span
              style={{
                color: last ? 'var(--hr-text)' : 'var(--hr-text-muted)',
                fontWeight: last ? 600 : 400,
              }}
            >
              {item.label}
            </span>
          )}
          {!last && <span style={{ color: 'var(--hr-text-dim)' }}>/</span>}
        </React.Fragment>
      );
    })}
  </nav>
);

/* ── StatCard ── */
export const StatCard: React.FC<{
  label: string;
  value: React.ReactNode;
  sub?: string;
  icon?: React.ReactNode;
  trend?: { value: number; label: string };
  color?: string;
}> = ({ label, value, sub, icon, trend, color = ACCENT }) => (
  <Card>
    <div
      style={{
        display: 'flex',
        alignItems: 'start',
        justifyContent: 'space-between',
        marginBottom: 8,
      }}
    >
      <span
        style={{
          fontSize: 11,
          fontWeight: 600,
          color: 'var(--hr-text-muted)',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
        }}
      >
        {label}
      </span>
      {icon && (
        <span style={{ padding: 6, borderRadius: 8, background: `${color}22`, color }}>{icon}</span>
      )}
    </div>
    <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--hr-text)' }}>{value}</div>
    {sub && <div style={{ fontSize: 11, color: 'var(--hr-text-muted)', marginTop: 4 }}>{sub}</div>}
    {trend && (
      <div
        style={{
          fontSize: 11,
          color: trend.value >= 0 ? 'var(--hr-status-approved)' : 'var(--hr-error)',
          marginTop: 4,
          fontWeight: 600,
        }}
      >
        {trend.value >= 0 ? '+' : ''}
        {trend.value}% {trend.label}
      </div>
    )}
  </Card>
);

/* ── Tabs ── */
export const Tabs: React.FC<{ tabs: string[]; active: string; onChange: (t: string) => void }> = ({
  tabs,
  active,
  onChange,
}) => (
  <div
    style={{
      display: 'flex',
      gap: 4,
      overflowX: 'auto',
      paddingBottom: 4,
      borderBottom: '1px solid var(--hr-border)',
      marginBottom: 16,
    }}
  >
    {tabs.map((tab) => (
      <button
        type="button"
        key={tab}
        onClick={() => onChange(tab)}
        style={{
          padding: '8px 14px',
          fontSize: 12,
          fontWeight: 600,
          background: 'transparent',
          border: 'none',
          borderBottom: active === tab ? `2px solid ${ACCENT}` : '2px solid transparent',
          color: active === tab ? 'var(--hr-text)' : 'var(--hr-text-muted)',
          cursor: 'pointer',
          flexShrink: 0,
          whiteSpace: 'nowrap',
        }}
      >
        {tab}
      </button>
    ))}
  </div>
);

/* ── Checkbox ── */
export const Checkbox: React.FC<{
  checked: boolean;
  onChange: (v: boolean) => void;
  label?: string;
}> = ({ checked, onChange, label }) => (
  <label
    style={{
      display: 'inline-flex',
      alignItems: 'center',
      gap: 8,
      cursor: 'pointer',
      fontSize: 12,
      color: 'var(--hr-text)',
    }}
  >
    <span
      onClick={() => onChange(!checked)}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        width: 16,
        height: 16,
        borderRadius: 4,
        border: `1px solid ${checked ? ACCENT : 'var(--hr-border)'}`,
        background: checked ? ACCENT : 'transparent',
        color: 'var(--hr-on-accent)',
        fontSize: 11,
        flexShrink: 0,
      }}
    >
      {checked ? '\u2713' : ''}
    </span>
    {label}
  </label>
);

/* Global keyframes style - injected once */
if (typeof document !== 'undefined' && !document.getElementById('hr-shimmer-style')) {
  const s = document.createElement('style');
  s.id = 'hr-shimmer-style';
  s.textContent = `@keyframes hr-shimmer { 0% { background-position: 200% 0; } 100% { background-position: -200% 0; } }`;
  document.head.appendChild(s);
}
