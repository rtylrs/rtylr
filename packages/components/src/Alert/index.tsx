import React from 'react';
import styles from './Alert.module.scss';

interface AlertProps {
  severity?: 'error' | 'warning' | 'info' | 'success';
  children: React.ReactNode;
  className?: string;
}

const icons: Record<NonNullable<AlertProps['severity']>, string> = {
  error: '✕',
  warning: '⚠',
  info: 'ℹ',
  success: '✓',
};

export const Alert: React.FC<AlertProps> = ({ severity = 'info', children, className }) => (
  <div
    role="alert"
    className={`${styles.alert} ${styles[severity]}${className ? ` ${className}` : ''}`}
  >
    <span className={styles.icon} aria-hidden="true">
      {icons[severity]}
    </span>
    <span>{children}</span>
  </div>
);

export default Alert;
