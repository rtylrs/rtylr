import React from 'react';
import styles from './Spinner.module.scss';

interface SpinnerProps {
  size?: number;
  className?: string;
}

export const Spinner: React.FC<SpinnerProps> = ({ size = 24, className }) => (
  <span
    className={`${styles.spinner}${className ? ` ${className}` : ''}`}
    style={{ width: size, height: size }}
    role="status"
    aria-label="Loading"
  />
);

export default Spinner;
