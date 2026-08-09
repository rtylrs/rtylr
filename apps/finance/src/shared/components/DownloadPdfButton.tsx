import React from 'react';
import { Download } from 'lucide-react';
import { ghostBtn } from '../lib/tokens';

interface DownloadPdfButtonProps {
  invoiceId: string;
  label?: string;
}

const DownloadPdfButton: React.FC<DownloadPdfButtonProps> = ({
  invoiceId: _invoiceId,
  label = 'Download PDF',
}) => (
  <button
    style={ghostBtn}
    aria-label="Download PDF"
    disabled
    title="PDF download coming in a future update"
  >
    <Download size={13} />
    {label}
  </button>
);

export default DownloadPdfButton;
