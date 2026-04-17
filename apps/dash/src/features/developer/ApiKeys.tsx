import React, { useState } from 'react';
import { KeyRound, Plus, Copy } from 'lucide-react';
import { toast } from 'sonner';
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
  Breadcrumbs,
  ConfirmDialog,
  EmptyState,
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
  { key: 'read', label: 'Read', description: 'Read access to all resources' },
  { key: 'write', label: 'Write', description: 'Create and update resources' },
  { key: 'admin', label: 'Admin', description: 'Full workspace access' },
];

const ApiKeys: React.FC = () => {
  const { data = [], isLoading } = useApiKeys();
  const create = useCreateApiKey();
  const revoke = useRevokeApiKey();

  const [createOpen, setCreateOpen] = useState(false);
  const [name, setName] = useState('');
  const [scopes, setScopes] = useState<string[]>(['read']);
  const [showKey, setShowKey] = useState<string | null>(null);
  const [toRevoke, setToRevoke] = useState<ApiKey | null>(null);

  const submitCreate = async () => {
    if (!name.trim()) return;
    const result = await create.mutateAsync({ name, scopes });
    setShowKey(result.key);
    setName('');
    setScopes(['read']);
    setCreateOpen(false);
  };

  return (
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
              title="No API keys yet"
              description="Create your first API key to start building."
              actionLabel="New key"
              onAction={() => setCreateOpen(true)}
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
              Copy it now — you won&apos;t be able to see it again.
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
              onClick={() => {
                if (showKey) {
                  navigator.clipboard.writeText(showKey);
                  toast.success('Key copied');
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
          if (toRevoke) await revoke.mutateAsync(toRevoke.id);
        }}
      />
    </div>
  );
};

export default ApiKeys;
