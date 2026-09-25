import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Webhook as WebhookIcon, Plus } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
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
      await create.mutateAsync({ url: normalizedUrl, events });
      setUrl('');
      setEvents([]);
      setCreateOpen(false);
      toast.success('Webhook created');
    } catch {
      // onError in hook handles toast
    }
  };

  return (
    <>
      <HelmetContainer title="Webhooks · rtylr" />
      <div className="space-y-5 max-w-4xl">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div
              className="h-9 w-9 rounded-xl flex items-center justify-center"
              style={{ background: 'rgba(69,131,196,0.15)' }}
            >
              <WebhookIcon size={16} style={{ color: '#4583C4' }} />
            </div>
            <div>
              <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
                Webhooks
              </h1>
              <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                Receive events in real time at your own endpoints.
              </p>
            </div>
          </div>
          <Button onClick={() => setCreateOpen(true)}>
            <Plus size={14} />
            New webhook
          </Button>
        </div>

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
