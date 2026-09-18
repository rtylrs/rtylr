import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { X, Loader2 } from 'lucide-react';

export const ACCENT = '#3b82f6';

// ── Card ─────────────────────────────────────────────────────────
export const Card: React.FC<
  React.PropsWithChildren<{ style?: React.CSSProperties; padded?: boolean }>
> = ({ children, style, padded = true }) => (
  <div
    style={{
      background: 'var(--hub-card)',
      border: '1px solid var(--hub-border)',
      borderRadius: '12px',
      padding: padded ? '16px' : 0,
      ...style,
    }}
  >
    {children}
  </div>
);

// ── PageHeader ───────────────────────────────────────────────────
export const PageHeader: React.FC<{
  title: string;
  subtitle?: string;
  actions?: React.ReactNode;
}> = ({ title, subtitle, actions }) => (
  <div
    style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      marginBottom: '20px',
    }}
  >
    <div>
      <h2 style={{ fontSize: '18px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>
        {title}
      </h2>
      {subtitle && (
        <p style={{ fontSize: '13px', color: 'var(--hub-text-muted)', margin: '2px 0 0' }}>
          {subtitle}
        </p>
      )}
    </div>
    {actions && <div style={{ display: 'flex', gap: '8px' }}>{actions}</div>}
  </div>
);

// ── Button ───────────────────────────────────────────────────────
export const Button: React.FC<
  React.ButtonHTMLAttributes<HTMLButtonElement> & {
    variant?: 'primary' | 'ghost' | 'danger' | 'outline';
  }
> = ({ variant = 'primary', style, children, ...props }) => {
  const base: React.CSSProperties = {
    padding: '8px 14px',
    borderRadius: '8px',
    fontSize: '13px',
    fontWeight: 600,
    cursor: 'pointer',
    border: '1px solid transparent',
    transition: 'all 0.15s',
    display: 'inline-flex',
    alignItems: 'center',
    gap: '6px',
    whiteSpace: 'nowrap',
  };
  const variants: Record<string, React.CSSProperties> = {
    primary: { background: ACCENT, color: '#fff' },
    ghost: { background: 'transparent', color: 'var(--hub-text-muted)' },
    outline: {
      background: 'transparent',
      color: 'var(--hub-text)',
      border: '1px solid var(--hub-border)',
    },
    danger: { background: '#ef4444', color: '#fff' },
  };
  return (
    <button type="button" {...props} style={{ ...base, ...variants[variant], ...style }}>
      {children}
    </button>
  );
};

// ── Input ────────────────────────────────────────────────────────
export const Input: React.FC<React.InputHTMLAttributes<HTMLInputElement> & { label?: string }> = ({
  label,
  style,
  ...props
}) => (
  <label style={{ display: 'flex', flexDirection: 'column', gap: '6px', width: '100%' }}>
    {label && (
      <span
        style={{
          fontSize: '11px',
          color: 'var(--hub-text-dim)',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
          fontWeight: 600,
        }}
      >
        {label}
      </span>
    )}
    <input
      {...props}
      style={{
        height: '36px',
        padding: '0 12px',
        borderRadius: '8px',
        background: 'var(--hub-input-bg)',
        border: '1px solid var(--hub-border)',
        color: 'var(--hub-text)',
        fontSize: '13px',
        outline: 'none',
        ...style,
      }}
    />
  </label>
);

// ── Textarea ─────────────────────────────────────────────────────
export const Textarea: React.FC<
  React.TextareaHTMLAttributes<HTMLTextAreaElement> & { label?: string }
> = ({ label, style, ...props }) => (
  <label style={{ display: 'flex', flexDirection: 'column', gap: '6px', width: '100%' }}>
    {label && (
      <span
        style={{
          fontSize: '11px',
          color: 'var(--hub-text-dim)',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
          fontWeight: 600,
        }}
      >
        {label}
      </span>
    )}
    <textarea
      {...props}
      style={{
        minHeight: '80px',
        padding: '10px 12px',
        borderRadius: '8px',
        background: 'var(--hub-input-bg)',
        border: '1px solid var(--hub-border)',
        color: 'var(--hub-text)',
        fontSize: '13px',
        outline: 'none',
        fontFamily: 'inherit',
        resize: 'vertical',
        ...style,
      }}
    />
  </label>
);

// ── Select ───────────────────────────────────────────────────────
export const Select: React.FC<
  React.SelectHTMLAttributes<HTMLSelectElement> & { label?: string }
> = ({ label, style, children, ...props }) => (
  <label style={{ display: 'flex', flexDirection: 'column', gap: '6px', width: '100%' }}>
    {label && (
      <span
        style={{
          fontSize: '11px',
          color: 'var(--hub-text-dim)',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
          fontWeight: 600,
        }}
      >
        {label}
      </span>
    )}
    <select
      {...props}
      style={{
        height: '36px',
        padding: '0 10px',
        borderRadius: '8px',
        background: 'var(--hub-input-bg)',
        border: '1px solid var(--hub-border)',
        color: 'var(--hub-text)',
        fontSize: '13px',
        outline: 'none',
        cursor: 'pointer',
        ...style,
      }}
    >
      {children}
    </select>
  </label>
);

// ── Modal ────────────────────────────────────────────────────────
export const Modal: React.FC<{
  open: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
  footer?: React.ReactNode;
  size?: 'sm' | 'md' | 'lg';
}> = ({ open, onClose, title, children, footer, size = 'md' }) => {
  const { t } = useTranslation();
  if (!open) return null;
  const widths = { sm: '380px', md: '520px', lg: '720px' };
  return (
    <div
      onClick={onClose}
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.5)',
        zIndex: 100,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '20px',
      }}
    >
      <motion.div
        initial={{ opacity: 0, scale: 0.96, y: 10 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        onClick={(e) => e.stopPropagation()}
        style={{
          width: '100%',
          maxWidth: widths[size],
          maxHeight: '90vh',
          background: 'var(--hub-card)',
          border: '1px solid var(--hub-border)',
          borderRadius: '14px',
          boxShadow: 'var(--hub-shadow)',
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '14px 18px',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          <h3 style={{ fontSize: '14px', fontWeight: 600, color: 'var(--hub-text)', margin: 0 }}>
            {title}
          </h3>
          <button
            aria-label={t('crm.components.closeDialog')}
            type="button"
            onClick={onClose}
            style={{
              background: 'transparent',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--hub-text-muted)',
              padding: '4px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <X size={16} />
          </button>
        </div>
        <div style={{ padding: '18px', overflowY: 'auto', flex: 1 }}>{children}</div>
        {footer && (
          <div
            style={{
              display: 'flex',
              justifyContent: 'flex-end',
              gap: '8px',
              padding: '12px 18px',
              borderTop: '1px solid var(--hub-border)',
            }}
          >
            {footer}
          </div>
        )}
      </motion.div>
    </div>
  );
};

// ── ConfirmModal ─────────────────────────────────────────────────
export const ConfirmModal: React.FC<{
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  description?: string;
  confirmLabel?: string;
  variant?: 'primary' | 'danger';
}> = ({ open, onClose, onConfirm, title, description, confirmLabel, variant = 'primary' }) => {
  const { t } = useTranslation();
  return (
    <Modal
      open={open}
      onClose={onClose}
      title={title}
      size="sm"
      footer={
        <>
          <Button variant="outline" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          <Button
            variant={variant}
            onClick={() => {
              onConfirm();
              onClose();
            }}
          >
            {confirmLabel ?? t('common.confirm')}
          </Button>
        </>
      }
    >
      {description && (
        <p style={{ fontSize: '13px', color: 'var(--hub-text-muted)', margin: 0, lineHeight: 1.6 }}>
          {description}
        </p>
      )}
    </Modal>
  );
};

// ── Skeleton ─────────────────────────────────────────────────────
export const Skeleton: React.FC<{
  width?: string | number;
  height?: string | number;
  style?: React.CSSProperties;
}> = ({ width = '100%', height = '14px', style }) => (
  <div
    style={{
      width,
      height,
      borderRadius: '6px',
      background:
        'linear-gradient(90deg, var(--hub-hover) 0%, var(--hub-active) 50%, var(--hub-hover) 100%)',
      backgroundSize: '200% 100%',
      animation: 'crm-skeleton 1.4s ease-in-out infinite',
      ...style,
    }}
  />
);

// ── SkeletonList ─────────────────────────────────────────────────
export const SkeletonList: React.FC<{ count?: number }> = ({ count = 4 }) => (
  <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
    {Array.from({ length: count }).map((_, i) => (
      <div key={i} style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
        <Skeleton width={32} height={32} style={{ borderRadius: '8px', flexShrink: 0 }} />
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '6px' }}>
          <Skeleton width="40%" height="12px" />
          <Skeleton width="70%" height="10px" />
        </div>
      </div>
    ))}
  </div>
);

// ── Breadcrumbs ──────────────────────────────────────────────────
export const Breadcrumbs: React.FC<{ items: Array<{ label: string; to?: string }> }> = ({
  items,
}) => (
  <nav
    style={{
      display: 'flex',
      alignItems: 'center',
      gap: '6px',
      fontSize: '12px',
      color: 'var(--hub-text-muted)',
      marginBottom: '12px',
    }}
  >
    {items.map((item, i) => (
      <React.Fragment key={i}>
        {item.to ? (
          <Link to={item.to} style={{ color: 'var(--hub-text-muted)', textDecoration: 'none' }}>
            {item.label}
          </Link>
        ) : (
          <span style={{ color: 'var(--hub-text)' }}>{item.label}</span>
        )}
        {i < items.length - 1 && <span style={{ color: 'var(--hub-text-dim)' }}>/</span>}
      </React.Fragment>
    ))}
  </nav>
);

// ── LoadingSpinner ───────────────────────────────────────────────
export const LoadingSpinner: React.FC<{ size?: number }> = ({ size = 16 }) => (
  <motion.span
    animate={{ rotate: 360 }}
    transition={{ repeat: Infinity, duration: 1, ease: 'linear' }}
    style={{ display: 'inline-flex' }}
  >
    <Loader2 size={size} />
  </motion.span>
);
