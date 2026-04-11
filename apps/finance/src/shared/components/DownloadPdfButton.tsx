import React from 'react';
import { Download } from 'lucide-react';
import { toast } from 'sonner';
import { ghostBtn } from '../lib/tokens';

interface DownloadPdfButtonProps {
  invoiceId: string;
  label?: string;
}

const DownloadPdfButton: React.FC<DownloadPdfButtonProps> = ({
  invoiceId,
  label = 'Download PDF',
}) => {
  const handle = () => {
    // Create a simple "fake" text blob as a placeholder for a PDF.
    const blob = new Blob(
      [
        `Invoice ${invoiceId}\nGenerated ${new Date().toLocaleString()}\n\nThis is a placeholder PDF export.`,
      ],
      { type: 'application/pdf' }
    );
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${invoiceId}.pdf`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
    toast.success(`Downloaded ${invoiceId}.pdf`);
  };

  return (
    <button onClick={handle} style={ghostBtn}>
      <Download size={13} />
      {label}
    </button>
  );
};

export default DownloadPdfButton;
