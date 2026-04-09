import React from 'react';
import { ArrowRight } from 'lucide-react';
import AuthButton, { AuthButtonProps } from './AuthButton';

export interface SubmitButtonProps extends AuthButtonProps {
  label: string;
  showArrow?: boolean;
}

/**
 * Canonical submit button with label + optional arrow + built-in loading.
 * Thin wrapper to reduce repetition across every screen.
 */
const SubmitButton: React.FC<SubmitButtonProps> = ({
  label,
  showArrow = true,
  ...props
}) => {
  return (
    <AuthButton type="submit" {...props}>
      <span className="inline-flex items-center justify-center gap-2">
        {label}
        {showArrow && <ArrowRight className="h-[18px] w-[18px]" />}
      </span>
    </AuthButton>
  );
};

export default SubmitButton;
