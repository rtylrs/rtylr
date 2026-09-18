import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { QrCode, Camera } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { Card, Button } from '../../shared/components/ui';

const QRScanner = (): React.JSX.Element => {
  const { t } = useTranslation();
  const [showPrompt, setShowPrompt] = useState(false);

  return (
    <Card style={{ padding: 20 }}>
      {showPrompt ? (
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            gap: 16,
            padding: 24,
          }}
        >
          <div
            style={{
              width: 200,
              height: 200,
              borderRadius: 12,
              background: 'var(--hr-hover)',
              border: '2px dashed var(--hr-border)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Camera size={40} style={{ color: 'var(--hr-text-dim)' }} />
          </div>
          <p style={{ fontSize: 12, color: 'var(--hr-text-muted)', textAlign: 'center' }}>
            {t('attendance.qrScanner.cameraPlaceholder')}
          </p>
          <Button variant="ghost" onClick={() => setShowPrompt(false)}>
            {t('common.cancel')}
          </Button>
        </div>
      ) : (
        <EmptyState
          icon={QrCode}
          title={t('attendance.qrScanner.title')}
          description={t('attendance.qrScanner.description')}
          action={{
            label: t('attendance.qrScanner.scanButton'),
            onClick: () => setShowPrompt(true),
          }}
        />
      )}
    </Card>
  );
};

export default QRScanner;
