import React from 'react';
import styles from './Chip.module.scss';

interface ChipProps {
  label: string;
  icon?: React.ReactNode;
  color?: 'primary' | 'secondary' | 'default';
  size?: 'sm' | 'md';
  className?: string;
}

export const Chip: React.FC<ChipProps> = ({
  label,
  icon,
  color = 'default',
  size = 'sm',
  className,
}) => (
  <span
    className={`${styles.chip} ${styles[color]} ${styles[size]}${className ? ` ${className}` : ''}`}
  >
    {icon && <span className={styles.icon}>{icon}</span>}
    {label}
  </span>
);

export default Chip;
