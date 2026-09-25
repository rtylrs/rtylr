import React from 'react';
import styles from './Switch.module.scss';

interface SwitchProps {
  checked: boolean;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  disabled?: boolean;
  className?: string;
}

export const Switch: React.FC<SwitchProps> = ({ checked, onChange, disabled, className }) => (
  <label className={`${styles.switch}${className ? ` ${className}` : ''}`}>
    <input type="checkbox" checked={checked} onChange={onChange} disabled={disabled} />
    <span className={styles.slider} />
  </label>
);

export default Switch;
