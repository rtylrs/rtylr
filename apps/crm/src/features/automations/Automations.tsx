import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import { useUserStore } from '@voxire/auth';
import { Zap, Plus, Play, Pause, Trash2, Cake, UserPlus, RotateCcw, Star } from 'lucide-react';
import {
  useListAutomations,
  useUpdateAutomation,
  useDeleteAutomation,
} from '../../shared/hooks/queries/useAutomations';
import {
  PageHeader,
  Card,
  Button,
  EmptyState,
  SkeletonList,
  ConfirmModal,
} from '../../shared/components/ui';
import { canCrmPermission } from '../../shared/lib/crmPermissions';

const Automations: React.FC = () => {
  const navigate = useNavigate();
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canManageAutomations = canCrmPermission(role, 'crm:automations');
  const { data: fetched, isLoading } = useListAutomations();
  const updateAutomation = useUpdateAutomation();
  const removeAutomation = useDeleteAutomation();
  const [confirmDeleteId, setConfirmDeleteId] = useState<number | null>(null);
  const automations = fetched ?? [];

  const handleToggleStatus = (id: number, current: string): void => {
    const next = current === 'active' ? 'paused' : 'active';
    updateAutomation.mutate(
      { id, data: { status: next as 'active' | 'paused', isActive: next === 'active' } },
      {
        onSuccess: () =>
          toast.success(next === 'active' ? 'Automation resumed' : 'Automation paused'),
        onError: () => toast.error('Failed to update automation'),
      }
    );
  };

  const handleDelete = (id: number): void => {
    removeAutomation.mutate(id, {
      onSuccess: () => {
        toast.success('Automation deleted');
        setConfirmDeleteId(null);
      },
      onError: () => {
        toast.error('Failed to delete automation');
        setConfirmDeleteId(null);
      },
    });
  };

  const deleteTarget = automations.find((a) => a.id === confirmDeleteId);

  const icons: Record<string, React.ComponentType<{ size?: number }>> = {
    'Welcome Series': UserPlus,
    'Birthday Greeting': Cake,
    'Win-back': RotateCcw,
    'Review Request': Star,
  };

  return (
    <>
      <HelmetContainer title="Automations · rtylr" />
      <div style={{ maxWidth: '1100px' }}>
        <PageHeader
          title="Automations"
          subtitle={`${automations.length} workflows configured`}
          actions={
            <Button variant="primary" onClick={() => navigate({ to: '/automations/new' })}>
              <Plus size={13} />
              New automation
            </Button>
          }
        />

        {isLoading ? (
          <Card>
            <SkeletonList count={4} />
          </Card>
        ) : automations.length === 0 ? (
          <Card>
            <EmptyState
              icon={Zap}
              title="No automations yet"
              description="Set up drip campaigns, birthday greetings, and win-back flows."
              action={
                <Button variant="primary" onClick={() => navigate({ to: '/automations/new' })}>
                  <Plus size={13} />
                  Create one
                </Button>
              }
            />
          </Card>
        ) : (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
              gap: '14px',
            }}
          >
            {automations.map((a) => {
              const Icon = icons[a.name] || Zap;
              return (
                <div
                  key={a.id}
                  onClick={() => navigate({ to: '/automations/new', search: { id: a.id } })}
                  style={{
                    background: 'var(--hub-card)',
                    border: '1px solid var(--hub-border)',
                    borderRadius: '14px',
                    padding: '16px',
                    cursor: 'pointer',
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '10px',
                    transition: 'border-color 0.15s',
                  }}
                  onMouseEnter={(e) => {
                    (e.currentTarget as HTMLElement).style.borderColor = 'rgba(59,130,246,0.4)';
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-border)';
                  }}
                >
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                    }}
                  >
                    <div
                      style={{
                        width: '38px',
                        height: '38px',
                        borderRadius: '10px',
                        background: 'rgba(59,130,246,0.12)',
                        color: '#3b82f6',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                      }}
                    >
                      <Icon size={18} />
                    </div>
                    <div
                      style={{ display: 'flex', alignItems: 'center', gap: '6px' }}
                      onClick={(e) => e.stopPropagation()}
                    >
                      <span
                        style={{
                          display: 'inline-flex',
                          alignItems: 'center',
                          gap: '4px',
                          padding: '3px 10px',
                          borderRadius: '9999px',
                          fontSize: '10px',
                          fontWeight: 600,
                          background:
                            a.status === 'active'
                              ? 'rgba(16,185,129,0.12)'
                              : a.status === 'paused'
                                ? 'rgba(245,158,11,0.12)'
                                : 'rgba(148,163,184,0.12)',
                          color:
                            a.status === 'active'
                              ? '#10b981'
                              : a.status === 'paused'
                                ? '#f59e0b'
                                : '#94a3b8',
                        }}
                      >
                        {a.status === 'active' ? (
                          <Play size={9} />
                        ) : a.status === 'paused' ? (
                          <Pause size={9} />
                        ) : null}
                        {a.status.toUpperCase()}
                      </span>
                      {a.status !== 'draft' && canManageAutomations && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleToggleStatus(a.id, a.status);
                          }}
                          disabled={updateAutomation.isPending}
                          aria-label={
                            a.status === 'active' ? 'Pause automation' : 'Resume automation'
                          }
                          style={{
                            padding: '4px 6px',
                            borderRadius: '6px',
                            background: 'transparent',
                            border: '1px solid var(--hub-border)',
                            color: 'var(--hub-text-muted)',
                            cursor: updateAutomation.isPending ? 'not-allowed' : 'pointer',
                            display: 'flex',
                            alignItems: 'center',
                            opacity: updateAutomation.isPending ? 0.6 : 1,
                          }}
                        >
                          {a.status === 'active' ? <Pause size={11} /> : <Play size={11} />}
                        </button>
                      )}
                      {canManageAutomations && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            setConfirmDeleteId(a.id);
                          }}
                          disabled={removeAutomation.isPending}
                          aria-label="Delete automation"
                          style={{
                            padding: '4px 6px',
                            borderRadius: '6px',
                            background: 'rgba(239,68,68,0.08)',
                            border: '1px solid rgba(239,68,68,0.25)',
                            color: '#ef4444',
                            cursor: removeAutomation.isPending ? 'not-allowed' : 'pointer',
                            display: 'flex',
                            alignItems: 'center',
                            opacity: removeAutomation.isPending ? 0.6 : 1,
                          }}
                        >
                          <Trash2 size={11} />
                        </button>
                      )}
                    </div>
                  </div>
                  <div>
                    <h3
                      style={{
                        fontSize: '14px',
                        fontWeight: 700,
                        color: 'var(--hub-text)',
                        margin: '0 0 2px',
                      }}
                    >
                      {a.name}
                    </h3>
                    <p style={{ fontSize: '11px', color: 'var(--hub-text-muted)', margin: 0 }}>
                      {a.description}
                    </p>
                  </div>
                  <div
                    style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      padding: '10px 12px',
                      background: 'var(--hub-hover)',
                      borderRadius: '8px',
                      fontSize: '11px',
                    }}
                  >
                    <div>
                      <div
                        style={{
                          color: 'var(--hub-text-dim)',
                          textTransform: 'uppercase',
                          fontWeight: 700,
                        }}
                      >
                        Trigger
                      </div>
                      <div style={{ color: 'var(--hub-text)', fontWeight: 600 }}>{a.trigger}</div>
                    </div>
                    <div>
                      <div
                        style={{
                          color: 'var(--hub-text-dim)',
                          textTransform: 'uppercase',
                          fontWeight: 700,
                        }}
                      >
                        Enrolled
                      </div>
                      <div style={{ color: 'var(--hub-text)', fontWeight: 600 }}>{a.enrolled}</div>
                    </div>
                    <div>
                      <div
                        style={{
                          color: 'var(--hub-text-dim)',
                          textTransform: 'uppercase',
                          fontWeight: 700,
                        }}
                      >
                        Done
                      </div>
                      <div style={{ color: 'var(--hub-text)', fontWeight: 600 }}>{a.completed}</div>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}

        <ConfirmModal
          open={confirmDeleteId !== null}
          onClose={() => setConfirmDeleteId(null)}
          onConfirm={() => confirmDeleteId !== null && handleDelete(confirmDeleteId)}
          title="Delete automation?"
          description={
            deleteTarget
              ? `Are you sure you want to delete "${deleteTarget.name}"? This action cannot be undone.`
              : ''
          }
          confirmLabel="Delete"
          variant="danger"
        />
      </div>
    </>
  );
};

export default Automations;
