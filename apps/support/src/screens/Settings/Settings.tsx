import React, { useState, useEffect } from 'react';
import { Settings as SettingsIcon, Save, Clock, Bell, Users, Shield } from 'lucide-react';
import { useSupportSettings, useUpdateSupportSettings } from '../../hooks/useSettings';
import { toast } from 'sonner';

const Settings: React.FC = () => {
  const { data: settings, isLoading, isError } = useSupportSettings();
  const updateSettings = useUpdateSupportSettings();

  const [autoAssign, setAutoAssign] = useState(false);
  const [roundRobin, setRoundRobin] = useState(false);
  const [slaEnabled, setSlaEnabled] = useState(false);
  const [slaFirstResponse, setSlaFirstResponse] = useState('');
  const [slaResolution, setSlaResolution] = useState('');
  const [businessHoursStart, setBusinessHoursStart] = useState('09:00');
  const [businessHoursEnd, setBusinessHoursEnd] = useState('17:00');
  const [businessDays, setBusinessDays] = useState('mon,tue,wed,thu,fri');
  const [emailNotifications, setEmailNotifications] = useState(true);
  const [slackNotifications, setSlackNotifications] = useState(false);

  useEffect(() => {
    if (settings) {
      setAutoAssign(settings.autoAssign ?? false);
      setRoundRobin(settings.roundRobin ?? false);
      setSlaEnabled(settings.slaEnabled ?? false);
      setSlaFirstResponse(settings.slaFirstResponseMinutes?.toString() ?? '');
      setSlaResolution(settings.slaResolutionMinutes?.toString() ?? '');
      setBusinessHoursStart(settings.businessHoursStart ?? '09:00');
      setBusinessHoursEnd(settings.businessHoursEnd ?? '17:00');
      setBusinessDays(settings.businessDays ?? 'mon,tue,wed,thu,fri');
      setEmailNotifications(settings.emailNotifications ?? true);
      setSlackNotifications(settings.slackNotifications ?? false);
    }
  }, [settings]);

  const handleSave = () => {
    updateSettings.mutate(
      {
        autoAssign,
        roundRobin,
        slaEnabled,
        slaFirstResponseMinutes: slaFirstResponse ? Number(slaFirstResponse) : null,
        slaResolutionMinutes: slaResolution ? Number(slaResolution) : null,
        businessHoursStart,
        businessHoursEnd,
        businessDays,
        emailNotifications,
        slackNotifications,
      },
      {
        onSuccess: () => toast.success('Settings saved'),
        onError: () => toast.error('Failed to save settings'),
      }
    );
  };

  if (isLoading) {
    return (
      <div className="space-y-4 max-w-2xl">
        {[...Array(4)].map((_, i) => (
          <div key={i} className="h-24 rounded-xl animate-pulse" style={{ background: 'var(--hub-hover)' }} />
        ))}
      </div>
    );
  }

  if (isError) {
    return (
      <div
        className="flex items-center justify-center py-16 rounded-xl text-sm"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)', color: 'var(--hub-text-muted)' }}
      >
        Failed to load settings. Please try again.
      </div>
    );
  }

  return (
    <div className="max-w-2xl">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>
            Settings
          </h2>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            Configure support workflows and preferences
          </p>
        </div>
        <button
          onClick={handleSave}
          disabled={updateSettings.isPending}
          className="flex items-center gap-2 h-8 px-4 rounded-lg text-xs font-medium text-white transition-colors disabled:opacity-40"
          style={{ background: '#4583C4' }}
        >
          <Save size={12} />
          Save changes
        </button>
      </div>

      <div className="space-y-4">
        {/* Auto-assignment */}
        <div
          className="rounded-xl p-5"
          style={{ border: '1px solid var(--hub-border)', background: 'var(--hub-card)' }}
        >
          <div className="flex items-center gap-2 mb-3">
            <Users size={14} style={{ color: '#4583C4' }} />
            <h3 className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
              Auto-assignment
            </h3>
          </div>
          <div className="space-y-3">
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={autoAssign}
                onChange={(e) => setAutoAssign(e.target.checked)}
                className="rounded"
              />
              <span className="text-sm" style={{ color: 'var(--hub-text)' }}>
                Automatically assign new tickets to agents
              </span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={roundRobin}
                onChange={(e) => setRoundRobin(e.target.checked)}
                disabled={!autoAssign}
                className="rounded"
              />
              <span className="text-sm" style={{ color: autoAssign ? 'var(--hub-text)' : 'var(--hub-text-dim)' }}>
                Use round-robin assignment
              </span>
            </label>
          </div>
        </div>

        {/* SLA */}
        <div
          className="rounded-xl p-5"
          style={{ border: '1px solid var(--hub-border)', background: 'var(--hub-card)' }}
        >
          <div className="flex items-center gap-2 mb-3">
            <Shield size={14} style={{ color: '#4583C4' }} />
            <h3 className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
              SLA Policies
            </h3>
          </div>
          <label className="flex items-center gap-3 cursor-pointer mb-3">
            <input
              type="checkbox"
              checked={slaEnabled}
              onChange={(e) => setSlaEnabled(e.target.checked)}
              className="rounded"
            />
            <span className="text-sm" style={{ color: 'var(--hub-text)' }}>
              Enable SLA tracking
            </span>
          </label>
          {slaEnabled && (
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs mb-1" style={{ color: 'var(--hub-text-muted)' }}>
                  First response (minutes)
                </label>
                <input
                  type="number"
                  value={slaFirstResponse}
                  onChange={(e) => setSlaFirstResponse(e.target.value)}
                  placeholder="60"
                  className="w-full h-9 px-3 rounded-lg text-sm outline-none"
                  style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
                />
              </div>
              <div>
                <label className="block text-xs mb-1" style={{ color: 'var(--hub-text-muted)' }}>
                  Resolution (minutes)
                </label>
                <input
                  type="number"
                  value={slaResolution}
                  onChange={(e) => setSlaResolution(e.target.value)}
                  placeholder="1440"
                  className="w-full h-9 px-3 rounded-lg text-sm outline-none"
                  style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
                />
              </div>
            </div>
          )}
        </div>

        {/* Business hours */}
        <div
          className="rounded-xl p-5"
          style={{ border: '1px solid var(--hub-border)', background: 'var(--hub-card)' }}
        >
          <div className="flex items-center gap-2 mb-3">
            <Clock size={14} style={{ color: '#4583C4' }} />
            <h3 className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
              Business Hours
            </h3>
          </div>
          <div className="grid grid-cols-2 gap-3 mb-3">
            <div>
              <label className="block text-xs mb-1" style={{ color: 'var(--hub-text-muted)' }}>Start</label>
              <input
                type="time"
                value={businessHoursStart}
                onChange={(e) => setBusinessHoursStart(e.target.value)}
                className="w-full h-9 px-3 rounded-lg text-sm outline-none"
                style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
              />
            </div>
            <div>
              <label className="block text-xs mb-1" style={{ color: 'var(--hub-text-muted)' }}>End</label>
              <input
                type="time"
                value={businessHoursEnd}
                onChange={(e) => setBusinessHoursEnd(e.target.value)}
                className="w-full h-9 px-3 rounded-lg text-sm outline-none"
                style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
              />
            </div>
          </div>
          <div>
            <label className="block text-xs mb-1" style={{ color: 'var(--hub-text-muted)' }}>
              Business days (comma-separated)
            </label>
            <input
              type="text"
              value={businessDays}
              onChange={(e) => setBusinessDays(e.target.value)}
              placeholder="mon,tue,wed,thu,fri"
              className="w-full h-9 px-3 rounded-lg text-sm outline-none"
              style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
            />
          </div>
        </div>

        {/* Notifications */}
        <div
          className="rounded-xl p-5"
          style={{ border: '1px solid var(--hub-border)', background: 'var(--hub-card)' }}
        >
          <div className="flex items-center gap-2 mb-3">
            <Bell size={14} style={{ color: '#4583C4' }} />
            <h3 className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
              Notifications
            </h3>
          </div>
          <div className="space-y-3">
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={emailNotifications}
                onChange={(e) => setEmailNotifications(e.target.checked)}
                className="rounded"
              />
              <span className="text-sm" style={{ color: 'var(--hub-text)' }}>
                Email notifications for new tickets
              </span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={slackNotifications}
                onChange={(e) => setSlackNotifications(e.target.checked)}
                className="rounded"
              />
              <span className="text-sm" style={{ color: 'var(--hub-text)' }}>
                Slack notifications
              </span>
            </label>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Settings;
