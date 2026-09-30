// Common style tokens used across the finance app.
import type { CSSProperties } from 'react';

export const ACCENT = 'var(--finance-accent)';
export const cardStyle: CSSProperties = {
  background: 'var(--hub-card)',
  border: '1px solid var(--hub-border)',
  borderRadius: 12,
};

export const inputStyle: CSSProperties = {
  padding: '8px 12px',
  borderRadius: 8,
  border: '1px solid var(--hub-border)',
  background: 'var(--hub-input-bg)',
  color: 'var(--hub-text)',
  fontSize: 13,
  outline: 'none',
  width: '100%',
  boxSizing: 'border-box',
};

export const primaryBtn: CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: 6,
  padding: '8px 16px',
  borderRadius: 8,
  background: ACCENT,
  color: 'var(--finance-white)',
  border: 'none',
  fontSize: 13,
  fontWeight: 600,
  cursor: 'pointer',
};

export const ghostBtn: CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: 6,
  padding: '8px 14px',
  borderRadius: 8,
  background: 'var(--hub-hover)',
  color: 'var(--hub-text-muted)',
  border: '1px solid var(--hub-border)',
  fontSize: 13,
  fontWeight: 500,
  cursor: 'pointer',
};

export const labelStyle: CSSProperties = {
  fontSize: 12,
  fontWeight: 600,
  color: 'var(--hub-text-muted)',
  letterSpacing: '0.03em',
};
