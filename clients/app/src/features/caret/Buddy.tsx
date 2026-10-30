import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { ArrowLeft } from 'lucide-react';
import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { BuddyServiceClient } from '@voxire/voxie/hub/caret/v1/client';
import type { BuddySettings, PutBuddySettingsRequest } from '@voxire/voxie/hub/caret/v1/types';
import './buddy.css';

const client = createOneKitClient(BuddyServiceClient, 'dash');

function SettingsForm({ settings, scope }: { settings: BuddySettings; scope: string }) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [form, setForm] = useState<PutBuddySettingsRequest>({
    enabled: settings.enabled ?? false,
    instructions: settings.instructions ?? '',
    intervalMinutes: settings.intervalMinutes ?? 15,
    dailyRunCap: settings.dailyRunCap ?? 24,
  });
  const save = useMutation({
    mutationFn: (input: PutBuddySettingsRequest) => client.putBuddySettings(input),
    onSuccess: (result) => {
      if (result.settings)
        queryClient.setQueryData(['caret-buddy', scope, 'settings'], result.settings);
      void queryClient.invalidateQueries({ queryKey: ['caret-buddy', scope, 'settings'] });
    },
  });
  return (
    <form
      className="buddy-settings"
      onSubmit={(event) => {
        event.preventDefault();
        save.mutate(form);
      }}
    >
      <div className="buddy-enable">
        <div>
          <strong>{t('common.caret.buddy.enabled')}</strong>
          <p>{t('common.caret.buddy.enabledHint')}</p>
        </div>
        <label className="buddy-switch">
          <input
            type="checkbox"
            checked={form.enabled}
            onChange={(event) => setForm({ ...form, enabled: event.target.checked })}
            aria-label={t('common.caret.buddy.enabled')}
          />
          <span />
        </label>
      </div>
      <label className="buddy-field">
        {t('common.caret.buddy.instructions')}
        <textarea
          rows={4}
          maxLength={65535}
          value={form.instructions}
          onChange={(event) => setForm({ ...form, instructions: event.target.value })}
          placeholder={t('common.caret.buddy.instructionsPlaceholder')}
        />
      </label>
      <div className="buddy-numbers">
        <label className="buddy-field">
          {t('common.caret.buddy.interval')}
          <input
            type="number"
            min="15"
            max="1440"
            step="1"
            required
            value={Number.isNaN(form.intervalMinutes) ? '' : form.intervalMinutes}
            onChange={(event) => setForm({ ...form, intervalMinutes: event.target.valueAsNumber })}
          />
        </label>
        <label className="buddy-field">
          {t('common.caret.buddy.cap')}
          <input
            type="number"
            min="1"
            max="96"
            step="1"
            required
            value={Number.isNaN(form.dailyRunCap) ? '' : form.dailyRunCap}
            onChange={(event) => setForm({ ...form, dailyRunCap: event.target.valueAsNumber })}
          />
        </label>
      </div>
      <p className="buddy-note">{t('common.caret.buddy.accountHint')}</p>
      {save.isError && (
        <p role="alert" className="buddy-error">
          {getOneKitErrorMessage(save.error, t('common.caret.buddy.saveFailed'))}
        </p>
      )}
      {save.isSuccess && <p role="status">{t('common.caret.buddy.saved')}</p>}
      <button className="buddy-primary" type="submit" disabled={save.isPending}>
        {save.isPending ? t('common.caret.buddy.saving') : t('common.save')}
      </button>
    </form>
  );
}

export default function Buddy({ onBack, scope }: { onBack: () => void; scope: string }) {
  const { t, i18n } = useTranslation();
  const queryClient = useQueryClient();
  const [decisionMessage, setDecisionMessage] = useState('');
  const settings = useQuery({
    queryKey: ['caret-buddy', scope, 'settings'],
    queryFn: async () => {
      const result = await client.getBuddySettings({});
      if (!result.settings) throw new Error(t('common.caret.buddy.loadFailed'));
      return result.settings;
    },
  });
  const runs = useQuery({
    queryKey: ['caret-buddy', scope, 'runs'],
    queryFn: async () => (await client.listBuddyRuns({})).runs ?? [],
  });
  const actions = useQuery({
    queryKey: ['caret-buddy', scope, 'actions'],
    queryFn: async () => (await client.listBuddyActions({ status: 'pending' })).actions ?? [],
  });
  const decide = useMutation({
    mutationFn: ({ actionId, decision }: { actionId: number; decision: 'approve' | 'deny' }) =>
      client.decideBuddyAction({ actionId, decision }),
    onMutate: () => setDecisionMessage(''),
    onSuccess: (result) => {
      setDecisionMessage(
        result.action?.status === 'failed'
          ? t('common.caret.buddy.executionFailed') +
              ' ' +
              JSON.stringify(result.action.result ?? '')
          : t('common.caret.buddy.decisionSaved')
      );
      void queryClient.invalidateQueries({ queryKey: ['caret-buddy', scope, 'actions'] });
      void queryClient.invalidateQueries({ queryKey: ['caret-buddy', scope, 'runs'] });
    },
  });
  const date = (value?: string) => {
    if (!value) return '';
    const parsed = new Date(value);
    return Number.isNaN(parsed.getTime()) ? value : parsed.toLocaleString(i18n.language);
  };
  return (
    <div className="buddy-page">
      <header className="buddy-header">
        <button
          type="button"
          className="portal-icon-button"
          onClick={onBack}
          aria-label={t('common.caret.buddy.back')}
        >
          <ArrowLeft size={18} />
        </button>
        <span>{t('common.caret.buddy.title')}</span>
      </header>
      <main className="buddy-content">
        <div className="buddy-intro">
          <span className="buddy-eyebrow">CARET / BUDDY</span>
          <h1>{t('common.caret.buddy.title')}</h1>
          <p>{t('common.caret.buddy.description')}</p>
        </div>
        <section className="buddy-section" aria-labelledby="buddy-settings-title">
          <div className="buddy-section-heading">
            <h2 id="buddy-settings-title">{t('common.caret.buddy.settings')}</h2>
            <p>{t('common.caret.buddy.settingsHint')}</p>
          </div>
          {settings.isPending ? (
            <p>{t('common.caret.buddy.loading')}</p>
          ) : settings.isError ? (
            <div role="alert">
              <p>{t('common.caret.buddy.loadFailed')}</p>
              <button type="button" onClick={() => void settings.refetch()}>
                {t('common.refresh')}
              </button>
            </div>
          ) : (
            <SettingsForm settings={settings.data} scope={scope} />
          )}
        </section>
        <section className="buddy-section" aria-labelledby="buddy-actions-title">
          <div className="buddy-section-heading">
            <h2 id="buddy-actions-title">{t('common.caret.buddy.actions')}</h2>
            <p>{t('common.caret.buddy.actionsHint')}</p>
          </div>
          {actions.isPending ? (
            <p>{t('common.caret.buddy.loading')}</p>
          ) : actions.isError ? (
            <div role="alert">
              <p>{t('common.caret.buddy.actionsFailed')}</p>
              <button type="button" onClick={() => void actions.refetch()}>
                {t('common.refresh')}
              </button>
            </div>
          ) : actions.data.length === 0 ? (
            <p className="buddy-empty">{t('common.caret.buddy.noActions')}</p>
          ) : (
            <div className="buddy-list">
              {actions.data.map((action) => (
                <article className="buddy-item" key={action.id}>
                  <div className="buddy-item-top">
                    <strong>{action.toolName}</strong>
                    <span>{date(action.createdAt)}</span>
                  </div>
                  <pre>{JSON.stringify(action.input, null, 2)}</pre>
                  {action.expiresAt && (
                    <p className="buddy-note">
                      {t('common.caret.buddy.expires', { date: date(action.expiresAt) })}
                    </p>
                  )}
                  <div className="buddy-buttons">
                    <button
                      className="buddy-primary"
                      type="button"
                      disabled={decide.isPending || action.id == null}
                      onClick={() =>
                        action.id != null &&
                        decide.mutate({ actionId: action.id, decision: 'approve' })
                      }
                    >
                      {t('common.caret.buddy.approve')}
                    </button>
                    <button
                      className="buddy-secondary"
                      type="button"
                      disabled={decide.isPending || action.id == null}
                      onClick={() =>
                        action.id != null &&
                        decide.mutate({ actionId: action.id, decision: 'deny' })
                      }
                    >
                      {t('common.caret.buddy.deny')}
                    </button>
                  </div>
                </article>
              ))}
            </div>
          )}
          {decisionMessage && <p role="status">{decisionMessage}</p>}
          {decide.isError && (
            <p role="alert" className="buddy-error">
              {getOneKitErrorMessage(decide.error, t('common.caret.buddy.decisionFailed'))}
            </p>
          )}
        </section>
        <section className="buddy-section" aria-labelledby="buddy-runs-title">
          <div className="buddy-section-heading">
            <h2 id="buddy-runs-title">{t('common.caret.buddy.runs')}</h2>
            <p>{t('common.caret.buddy.runsHint')}</p>
          </div>
          {runs.isPending ? (
            <p>{t('common.caret.buddy.loading')}</p>
          ) : runs.isError ? (
            <div role="alert">
              <p>{t('common.caret.buddy.runsFailed')}</p>
              <button type="button" onClick={() => void runs.refetch()}>
                {t('common.refresh')}
              </button>
            </div>
          ) : runs.data.length === 0 ? (
            <p className="buddy-empty">{t('common.caret.buddy.noRuns')}</p>
          ) : (
            <div className="buddy-list">
              {runs.data.map((run) => (
                <article className="buddy-item" key={run.id}>
                  <div className="buddy-item-top">
                    <strong>{run.status}</strong>
                    <span>{date(run.createdAt)}</span>
                  </div>
                  {run.report && <p className="buddy-report">{run.report}</p>}
                  {run.error && <p className="buddy-error">{run.error}</p>}
                </article>
              ))}
            </div>
          )}
        </section>
      </main>
    </div>
  );
}
