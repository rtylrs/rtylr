import React from 'react';
import { Printer } from 'lucide-react';

interface PrintButtonProps {
  label?: string;
  targetId?: string;
  className?: string;
}

const PrintButton: React.FC<PrintButtonProps> = ({
  label = 'Print Receipt',
  targetId,
  className = '',
}) => {
  const handlePrint = () => {
    // If a targetId is provided, scroll the element into view first for window.print().
    if (targetId) {
      document.getElementById(targetId)?.scrollIntoView();
    }
    // window.print is available in Tauri webviews
    if (typeof window !== 'undefined' && typeof window.print === 'function') {
      window.print();
    }
  };
  return (
    <button
      type="button"
      onClick={handlePrint}
      className={`inline-flex h-12 items-center justify-center gap-2 px-5 rounded-lg bg-pos-text text-pos-bg font-bold text-sm hover:opacity-90 transition ${className}`}
    >
      <Printer className="h-4 w-4" aria-hidden="true" strokeWidth={2} />
      {label}
    </button>
  );
};

export default PrintButton;
