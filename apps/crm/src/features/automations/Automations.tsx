import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Zap, Plus, Play, Pause, Cake, UserPlus, RotateCcw, Star } from 'lucide-react';
import { useListAutomations } from '../../shared/hooks/queries/useAutomations';
import { PageHeader, Card, Button, EmptyState, SkeletonList } from '../../shared/components/ui';

const Automations: React.FC = () => {
  const navigate = useNavigate();
  const { data: fetched, isLoading } = useListAutomations();
  const automations = fetched ?? [];

  const icons: Record<string, React.ComponentType<{ size?: number }>> = {
    'Welcome Series': UserPlus,
    'Birthday Greeting': Cake,
    'Win-back': RotateCcw,
    'Review Request': Star,
  };

  return (
    <div style={{ maxWidth: '1100px' }}>
      <PageHeader
        title="Automations"
        subtitle={`${automations.length} workflows configured`}
        actions={
          <Button variant="primary" onClick={() => navigate('/automations/new')}>
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
              <Button variant="primary" onClick={() => navigate('/automations/new')}>
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
                onClick={() => navigate(`/automations/${a.id}`)}
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
                  style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}
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
    </div>
  );
};

export default Automations;
