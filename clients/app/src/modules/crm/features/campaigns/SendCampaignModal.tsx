import React from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Modal } from '../../shared/components/ui';
import { useCampaignAudience, useSendCampaign } from '../../shared/hooks/queries/useCampaigns';

/**
 * Confirms sending an email campaign, showing who it will actually reach:
 * only customers with an email address who agreed to marketing email.
 */
const SendCampaignModal: React.FC<{
  open: boolean;
  onClose: () => void;
  campaignId: number | undefined;
  segmentId: number | undefined;
  onSent?: () => void;
}> = ({ open, onClose, campaignId, segmentId, onSent }) => {
  const { t } = useTranslation();
  const audience = useCampaignAudience(segmentId, open);
  const send = useSendCampaign();
  const eligible = audience.data?.eligible ?? 0;
  const total = audience.data?.total ?? 0;

  const confirm = (): void => {
    if (!campaignId) return;
    send.mutate(campaignId, {
      onSuccess: () => {
        onClose();
        onSent?.();
      },
    });
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t('crm.campaignSend.title')}
      size="sm"
      footer={
        <>
          <Button variant="outline" onClick={onClose}>
            {t('crmCommon.cancel')}
          </Button>
          <Button
            variant="primary"
            onClick={confirm}
            disabled={!campaignId || send.isPending || audience.isPending || eligible === 0}
          >
            {send.isPending
              ? t('crm.campaignSend.sending')
              : t('crm.campaignSend.confirm', { count: eligible })}
          </Button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: '10px', fontSize: '13px' }}>
        {audience.isPending ? (
          <p style={{ margin: 0, color: 'var(--hub-text-muted)' }}>{t('crmCommon.loading')}</p>
        ) : audience.isError ? (
          <p style={{ margin: 0, color: '#ef4444' }}>{t('crm.campaignSend.audienceFailed')}</p>
        ) : (
          <p style={{ margin: 0, color: 'var(--hub-text)' }}>
            {t('crm.campaignSend.reach', { count: eligible, total })}
          </p>
        )}
        {!audience.isPending && !audience.isError && eligible === 0 && (
          <p style={{ margin: 0, color: 'var(--hub-text-muted)' }}>
            {t('crm.campaignSend.nobody')}
          </p>
        )}
        <p style={{ margin: 0, color: 'var(--hub-text-muted)', lineHeight: 1.5 }}>
          {t('crm.campaignSend.footerNote')}
        </p>
      </div>
    </Modal>
  );
};

export default SendCampaignModal;
