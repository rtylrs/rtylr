import { PageHeader } from '../../shared/components/dash/PageHeader';
import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Webhook as WebhookIcon, Plus, Copy } from 'lucide-react';
import { EmptyState } from '../../shared/components/dash/EmptyState';
import {
  Button,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  Input,
  Label,
} from '../../shared/components/ui';
import {
  ConfirmDialog,
  PermissionCheckbox,
  SkeletonList,
  WebhookRow,
} from '../../shared/components/dash';
import {
  useWebhooks,
  useCreateWebhook,
  useDeleteWebhook,
  type Webhook,
} from '../../shared/hooks/queries/useWebhooks';
import { normalizeWebhookUrl, SUPPORTED_WEBHOOK_EVENTS } from '../../shared/lib/webhookUrl';

const EVENTS = SUPPORTED_WEBHOOK_EVENTS;

const Webhooks: React.FC = () => {
  const { t } = useTranslation();
  const { data = [], isLoading } = useWebhooks();
  const create = useCreateWebhook();
  const remove = useDeleteWebhook();
  const [createOpen, setCreateOpen] = useState(false);
  const [url, setUrl] = useState('');
  const [events, setEvents] = useState<string[]>([]);
  const [toDelete, setToDelete] = useState<Webhook | null>(null);
  const [showSecret, setShowSecret] = useState<string | null>(null);

  const submitCreate = async () => {
    const normalizedUrl = normalizeWebhookUrl(url);
    if (!normalizedUrl) {
      toast.error(
        url.trim() ? 'Webhook URL must be a valid HTTP(S) URL' : 'Webhook URL is required'
      );
      return;
    }
    if (events.length === 0) {
      toast.error('Select at least one event');
      return;
    }
    try {
      const created = await create.mutateAsync({ url: normalizedUrl, events });
      setUrl('');
      setEvents([]);
      setCreateOpen(false);
      if (created?.secret) setShowSecret(created.secret);
    } catch {
      // onError in hook handles toast
    }
  };

  return (
    <>
      <HelmetContainer title="Webhooks · rtylr" />
      <div className="portal-management-page space-y-5 max-w-4xl">
        <PageHeader
          title={t('common.nav.webhooks')}
          description={t('common.management.webhooksDescription')}
          actions={
            <Button onClick={() => setCreateOpen(true)}>
              <Plus />
              {t('common.management.newWebhook')}
            </Button>
          }
        />

        <div
          className="rounded-xl overflow-hidden"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          {isLoading ? (
            <div className="p-2">
              <SkeletonList rows={3} withAvatar={false} />
            </div>
          ) : data.length === 0 ? (
            <div className="p-4">
              <EmptyState
                icon={WebhookIcon}
                title={t('developer.webhooks.emptyTitle')}
                description={t('developer.webhooks.emptyDescription')}
                action={{
                  label: t('developer.webhooks.newWebhook'),
                  onClick: () => setCreateOpen(true),
                }}
              />
            </div>
          ) : (
            <div className="p-2">
              {data.map((w) => (
                <WebhookRow
                  key={w.id}
                  webhook={w}
                  onDelete={(id) => setToDelete(data.find((x) => x.id === id) ?? null)}
                />
              ))}
            </div>
          )}
        </div>

        <Dialog open={createOpen} onOpenChange={setCreateOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>New webhook</DialogTitle>
              <DialogDescription>
                We&apos;ll POST JSON to this URL for the selected events.
              </DialogDescription>
            </DialogHeader>
            <div className="px-6 space-y-4">
              <div className="space-y-1.5">
                <Label htmlFor="wh-url">Endpoint URL</Label>
                <Input
                  id="wh-url"
                  type="url"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  placeholder="https://api.example.com/hooks"
                />
              </div>
              <div className="space-y-2">
                <Label>Events</Label>
                <div className="grid grid-cols-2 gap-2">
                  {EVENTS.map((ev) => (
                    <label
                      key={ev}
                      className="flex items-center gap-2 px-2 py-1.5 rounded-lg cursor-pointer transition-colors hover:bg-[var(--dash-hover)]"
                    >
                      <PermissionCheckbox
                        checked={events.includes(ev)}
                        onChange={(v) =>
                          setEvents((prev) => (v ? [...prev, ev] : prev.filter((e) => e !== ev)))
                        }
                      />
                      <code
                        className="text-xs font-mono"
                        style={{ color: 'var(--dash-text-muted)' }}
                      >
                        {ev}
                      </code>
                    </label>
                  ))}
                </div>
              </div>
            </div>
            <DialogFooter>
              <Button variant="secondary" onClick={() => setCreateOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={submitCreate}
                disabled={create.isPending || !url.trim() || events.length === 0}
              >
                Create webhook
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <Dialog open={!!showSecret} onOpenChange={(o) => !o && setShowSecret(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{t('developer.webhooks.secretTitle')}</DialogTitle>
              <DialogDescription>{t('developer.webhooks.secretDescription')}</DialogDescription>
            </DialogHeader>
            <div className="px-6 pb-4">
              <div
                className="flex items-center gap-2 p-3 rounded-lg font-mono text-xs break-all"
                style={{
                  background: 'var(--dash-hover)',
                  border: '1px solid var(--dash-border)',
                  color: 'var(--dash-text)',
                }}
                dir="ltr"
              >
                {showSecret}
              </div>
            </div>
            <DialogFooter>
              <Button
                onClick={async () => {
                  if (!showSecret) return;
                  try {
                    await navigator.clipboard.writeText(showSecret);
                    toast.success(t('developer.webhooks.secretCopied'));
                  } catch {
                    toast.error(t('developer.webhooks.secretCopyFailed'));
                  }
                }}
              >
                <Copy size={13} />
                {t('developer.webhooks.copySecret')}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <ConfirmDialog
          open={!!toDelete}
          onOpenChange={(o) => !o && setToDelete(null)}
          title="Delete webhook?"
          description="This endpoint will stop receiving events immediately."
          confirmLabel="Delete"
          destructive
          isLoading={remove.isPending}
          onConfirm={async () => {
            if (toDelete)
              try {
                await remove.mutateAsync(toDelete.id);
              } catch {
                // onError in hook handles toast
              }
          }}
        />
      </div>
    </>
  );
};

export default Webhooks;
