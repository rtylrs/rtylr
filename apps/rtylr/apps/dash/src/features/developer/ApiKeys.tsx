import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { KeyRound, Plus, Copy } from 'lucide-react';
import { toast } from 'sonner';
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
  ApiKeyRow,
  ConfirmDialog,
  PermissionCheckbox,
  SkeletonList,
} from '../../shared/components/dash';
import {
  useApiKeys,
  useCreateApiKey,
  useRevokeApiKey,
  type ApiKey,
} from '../../shared/hooks/queries/useApiKeys';

const SCOPE_OPTIONS = [
  {
    key: 'upload',
    label: 'Upload (CDN)',
    description: 'Upload and serve media via the CDN (cdn.voxire.com)',
  },
  { key: 'flow:read', label: 'Flow · Read', description: 'Read workspaces, projects, and tasks' },
  { key: 'flow:write', label: 'Flow · Write', description: 'Create and update projects and tasks' },
  { key: 'crm:read', label: 'CRM · Read', description: 'Read customers, campaigns, and loyalty' },
  { key: 'crm:write', label: 'CRM · Write', description: 'Create and update CRM records' },
  {
    key: 'finance:read',
    label: 'Finance · Read',
    description: 'Read invoices, expenses, and reports',
  },
  { key: 'finance:write', label: 'Finance · Write', description: 'Create invoices and expenses' },
  { key: 'erp:read', label: 'ERP · Read', description: 'Read inventory, products, and suppliers' },
  {
    key: 'erp:write',
    label: 'ERP · Write',
    description: 'Create and update inventory and products',
  },
];

const ApiKeys: React.FC = () => {
  const { t } = useTranslation();
  const { data = [], isLoading } = useApiKeys();
  const create = useCreateApiKey();
  const revoke = useRevokeApiKey();

  const [createOpen, setCreateOpen] = useState(false);
  const [name, setName] = useState('');
  const [scopes, setScopes] = useState<string[]>(['upload']);
  const [showKey, setShowKey] = useState<string | null>(null);
  const [toRevoke, setToRevoke] = useState<ApiKey | null>(null);

  const submitCreate = async () => {
    if (!name.trim()) {
      toast.error('API key name is required');
      return;
    }
    try {
      const result = await create.mutateAsync({ name, scopes });
      if (result.key) {
        setShowKey(result.key);
      } else {
        toast.success('API key created');
      }
      setName('');
      setScopes(['upload']);
      setCreateOpen(false);
    } catch {
      // onError in hook handles toast
    }
  };

  return (
    <>
      <HelmetContainer title="API keys · rtylr" />
      <div className="space-y-5 max-w-4xl">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div
              className="h-9 w-9 rounded-xl flex items-center justify-center"
              style={{ background: 'rgba(69,131,196,0.15)' }}
            >
              <KeyRound size={16} style={{ color: '#4583C4' }} />
            </div>
            <div>
              <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
                API keys
              </h1>
              <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                Keys grant programmatic access to your workspace data.
              </p>
            </div>
          </div>
          <Button onClick={() => setCreateOpen(true)}>
            <Plus size={14} />
            New key
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
                icon={KeyRound}
                title={t('developer.apiKeys.emptyTitle')}
                description={t('developer.apiKeys.emptyDescription')}
                action={{
                  label: t('developer.apiKeys.newKey'),
                  onClick: () => setCreateOpen(true),
                }}
              />
            </div>
          ) : (
            <div className="p-2">
              {data.map((k) => (
                <ApiKeyRow
                  key={k.id}
                  apiKey={k}
                  onRevoke={(id) => setToRevoke(data.find((x) => x.id === id) ?? null)}
                />
              ))}
            </div>
          )}
        </div>

        <Dialog open={createOpen} onOpenChange={setCreateOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create API key</DialogTitle>
              <DialogDescription>Give your key a descriptive name and scopes.</DialogDescription>
            </DialogHeader>
            <div className="px-6 space-y-4">
              <div className="space-y-1.5">
                <Label htmlFor="key-name">Name</Label>
                <Input
                  id="key-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Production server"
                />
              </div>
              <div className="space-y-2">
                <Label>Scopes</Label>
                {SCOPE_OPTIONS.map((opt) => (
                  <label
                    key={opt.key}
                    className="flex items-start gap-3 px-3 py-2 rounded-lg cursor-pointer transition-colors hover:bg-[var(--dash-hover)]"
                  >
                    <PermissionCheckbox
                      checked={scopes.includes(opt.key)}
                      onChange={(v) =>
                        setScopes((prev) =>
                          v ? [...prev, opt.key] : prev.filter((s) => s !== opt.key)
                        )
                      }
                    />
                    <div>
                      <p className="text-xs font-semibold" style={{ color: 'var(--dash-text)' }}>
                        {opt.label}
                      </p>
                      <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                        {opt.description}
                      </p>
                    </div>
                  </label>
                ))}
              </div>
            </div>
            <DialogFooter>
              <Button variant="secondary" onClick={() => setCreateOpen(false)}>
                Cancel
              </Button>
              <Button onClick={submitCreate} disabled={create.isPending || !name.trim()}>
                Create key
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <Dialog open={!!showKey} onOpenChange={(o) => !o && setShowKey(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Your new API key</DialogTitle>
              <DialogDescription>
                Copy it now, you won&apos;t be able to see it again.
              </DialogDescription>
            </DialogHeader>
            <div className="px-6 pb-4">
              <div
                className="flex items-center gap-2 p-3 rounded-lg font-mono text-xs break-all"
                style={{
                  background: 'var(--dash-hover)',
                  border: '1px solid var(--dash-border)',
                  color: 'var(--dash-text)',
                }}
              >
                {showKey}
              </div>
            </div>
            <DialogFooter>
              <Button
                onClick={async () => {
                  if (showKey) {
                    try {
                      await navigator.clipboard.writeText(showKey);
                      toast.success(t('developer.apiKeys.copied', 'Key copied'));
                    } catch {
                      toast.error(
                        t('developer.apiKeys.copyFailed', 'Failed to copy key to clipboard')
                      );
                    }
                  }
                }}
              >
                <Copy size={13} />
                Copy key
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <ConfirmDialog
          open={!!toRevoke}
          onOpenChange={(o) => !o && setToRevoke(null)}
          title="Revoke API key?"
          description={`The key "${toRevoke?.name}" will stop working immediately. This cannot be undone.`}
          confirmLabel="Revoke"
          destructive
          isLoading={revoke.isPending}
          onConfirm={async () => {
            if (toRevoke)
              try {
                await revoke.mutateAsync(toRevoke.id);
              } catch {
                // onError in hook handles toast
              }
          }}
        />
      </div>
    </>
  );
};

export default ApiKeys;
