import React from 'react';

interface WizardLayoutProps {
  children: React.ReactNode;
}

const WizardLayout = ({ children }: WizardLayoutProps): React.JSX.Element => {
  return (
    <div
      className="flex min-h-screen items-center justify-center"
      style={{ background: 'var(--hr-bg)' }}
    >
      <div className="w-full max-w-2xl px-4 py-8 sm:py-12">{children}</div>
    </div>
  );
};

export default WizardLayout;
