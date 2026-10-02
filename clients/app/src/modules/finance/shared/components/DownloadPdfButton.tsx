import React from 'react';
import { useTranslation } from 'react-i18next';
import { Download } from 'lucide-react';
import { ghostBtn } from '../lib/tokens';

interface DownloadPdfButtonProps {
  invoiceId: string;
  label?: string;
}

const DownloadPdfButton: React.FC<DownloadPdfButtonProps> = ({ invoiceId: _invoiceId, label }) => {
  const { t } = useTranslation('financeCommon');
  return (
    <button
      type="button"
      style={ghostBtn}
      aria-label={t('finance.pdf.download')}
      disabled
      title={t('finance.pdf.comingSoon')}
    >
      <Download size={13} />
      {label ?? t('finance.pdf.download')}
    </button>
  );
};

export default DownloadPdfButton;
