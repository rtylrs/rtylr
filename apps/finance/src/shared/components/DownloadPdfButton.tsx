import React from 'react';
import { Download } from 'lucide-react';
import { toast } from 'sonner';
import { ghostBtn } from '../lib/tokens';

interface DownloadPdfButtonProps {
  invoiceId: string;
  label?: string;
}

const DownloadPdfButton: React.FC<DownloadPdfButtonProps> = ({
  invoiceId: _invoiceId,
  label = 'Download PDF',
}) => {
  const handle = (): void => {
    toast.info('PDF download coming soon');
  };

  return (
    <button onClick={handle} style={ghostBtn} aria-label="Download PDF">
      <Download size={13} />
      {label}
    </button>
  );
};

export default DownloadPdfButton;
