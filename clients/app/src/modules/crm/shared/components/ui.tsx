import React from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '../../../../shared/components/ui/dialog';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { Loader2 } from 'lucide-react';

export const ACCENT = 'var(--crm-accent)';

// ── Card ─────────────────────────────────────────────────────────
export const Card: React.FC<
  React.PropsWithChildren<{ style?: React.CSSProperties; padded?: boolean }>
> = ({ children, style, padded = true }) => (
  <div
    style={{
      background: 'var(--hub-card)',
      border: '1px solid var(--hub-border)',
      borderRadius: '10px',
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
      <h2 style={{ fontSize: '26px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>
        {title}
      </h2>
      {subtitle && (
        <p style={{ fontSize: '13px', color: 'var(--hub-text-muted)', margin: '6px 0 0' }}>
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
    minHeight: '38px',
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
    primary: { background: 'var(--dash-active)', color: 'var(--dash-text)' },
    ghost: { background: 'transparent', color: 'var(--hub-text-muted)' },
    outline: {
      background: 'transparent',
      color: 'var(--hub-text)',
      border: '1px solid var(--hub-border)',
    },
    danger: { background: '#ef4444', color: '#fff' },
  };
  return (
    <button
      type="button"
      {...props}
      className="portal-crm-button"
      style={{ ...base, ...variants[variant], ...style }}
    >
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
          fontSize: '12px',
          color: 'var(--hub-text-muted)',
          fontWeight: 600,
        }}
      >
        {label}
      </span>
    )}
    <input
      {...props}
      style={{
        height: '40px',
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
          fontSize: '12px',
          color: 'var(--hub-text-muted)',
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
          fontSize: '12px',
          color: 'var(--hub-text-muted)',
          fontWeight: 600,
        }}
      >
        {label}
      </span>
    )}
    <select
      {...props}
      style={{
        height: '40px',
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
  const widths = { sm: '400px', md: '560px', lg: '760px' };
  return (
    <Dialog
      open={open}
      onOpenChange={(value) => {
        if (!value) onClose();
      }}
    >
      <DialogContent aria-describedby={undefined} style={{ maxWidth: widths[size] }}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <div className="px-6 pb-6 overflow-y-auto">{children}</div>
        {footer && <DialogFooter>{footer}</DialogFooter>}
      </DialogContent>
    </Dialog>
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
            {t('crmCommon.cancel')}
          </Button>
          <Button
            variant={variant}
            onClick={() => {
              onConfirm();
              onClose();
            }}
          >
            {confirmLabel ?? t('crmCommon.confirm')}
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
        {i < items.length - 1 && <span style={{ color: 'var(--hub-text-muted)' }}>/</span>}
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
