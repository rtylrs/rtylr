import { useEffect, useRef, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import type { TFunction } from 'i18next';
import { Trash2, X } from 'lucide-react';
import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { BuddyServiceClient } from '@voxire/voxie/hub/caret/v1/client';
import type {
  BuddyRoutine,
  BuddySettings,
  PutBuddySettingsRequest,
} from '@voxire/voxie/hub/caret/v1/types';
import BuddySprite from './BuddySprite';
import { BUDDY_COLORS, useBuddyColor, type BuddyColor } from './useBuddyColor';
import './buddy.css';

export const buddyClient = createOneKitClient(BuddyServiceClient, 'dash');

// Sent with every message in the Buddy chat so Caret answers as the bot.
export const BUDDY_SYSTEM_CONTEXT = `You are Caret Buddy, this person's personal assistant bot inside rtylr. This chat is your one ongoing conversation with them.
- Be proactive and brief, like a capable assistant texting.
- When they want something done regularly (check-ins, reminders, summaries, watching tickets, nudges), set up a routine with buddy.create_routine. Pick a short name, write instructions for your future self, choose every_minutes (60 hourly, 1440 daily, 10080 weekly) and first_run_at in their local time with its offset.
- To pause, resume, change or remove one, call buddy.routines for the id first, then buddy.update_routine or buddy.delete_routine.
- Routine results are posted back into this chat, so treat earlier routine messages as your own.`;

function every(minutes: number | undefined, t: TFunction) {
  const value = minutes ?? 0;
  if (value % 10080 === 0) return t('common.caret.buddy.everyWeeks', { count: value / 10080 });
  if (value % 1440 === 0) return t('common.caret.buddy.everyDays', { count: value / 1440 });
  if (value % 60 === 0) return t('common.caret.buddy.everyHours', { count: value / 60 });
  return t('common.caret.buddy.everyMinutes', { count: value });
}

function SupportAlerts({ settings, scope }: { settings: BuddySettings; scope: string }) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [form, setForm] = useState<PutBuddySettingsRequest>({
    enabled: settings.enabled ?? false,
    instructions: settings.instructions ?? '',
    intervalMinutes: settings.intervalMinutes || 60,
    dailyRunCap: settings.dailyRunCap || 24,
  });
  const save = useMutation({
    mutationFn: (input: PutBuddySettingsRequest) => buddyClient.putBuddySettings(input),
    onSuccess: (result) => {
      if (result.settings)
        queryClient.setQueryData(['caret-buddy', scope, 'settings'], result.settings);
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
          <strong>{t('common.caret.buddy.alerts')}</strong>
          <p>{t('common.caret.buddy.alertsHint')}</p>
        </div>
        <label className="buddy-switch">
          <input
            type="checkbox"
            checked={form.enabled}
            onChange={(event) => setForm({ ...form, enabled: event.target.checked })}
            aria-label={t('common.caret.buddy.alerts')}
          />
          <span />
        </label>
      </div>
      <label className="buddy-field">
        {t('common.caret.buddy.instructions')}
        <textarea
          rows={3}
          maxLength={65535}
          value={form.instructions}
          onChange={(event) => setForm({ ...form, instructions: event.target.value })}
          placeholder={t('common.caret.buddy.instructionsPlaceholder')}
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

// BuddyRoutines is the drawer beside the Buddy chat: the person's routines,
// and for admins the business-wide support alerts.
export default function BuddyRoutines({
  scope,
  isAdmin,
  onClose,
}: {
  scope: string;
  isAdmin: boolean;
  onClose: () => void;
}) {
  const { t, i18n } = useTranslation();
  const queryClient = useQueryClient();
  const closeRef = useRef<HTMLButtonElement>(null);
  const routinesKey = ['caret-buddy', scope, 'routines'];
  const [color, setColor] = useBuddyColor(scope);
  const routines = useQuery({
    queryKey: routinesKey,
    queryFn: async () => (await buddyClient.listBuddyRoutines({})).routines ?? [],
  });
  const settings = useQuery({
    queryKey: ['caret-buddy', scope, 'settings'],
    enabled: isAdmin,
    queryFn: async () => (await buddyClient.getBuddySettings({})).settings ?? {},
  });
  const toggle = useMutation({
    mutationFn: (routine: BuddyRoutine) =>
      buddyClient.setBuddyRoutineEnabled({ id: routine.id ?? 0, enabled: !routine.enabled }),
    onSettled: () => void queryClient.invalidateQueries({ queryKey: routinesKey }),
  });
  const remove = useMutation({
    mutationFn: (routine: BuddyRoutine) => buddyClient.deleteBuddyRoutine({ id: routine.id ?? 0 }),
    onSettled: () => void queryClient.invalidateQueries({ queryKey: routinesKey }),
  });
  useEffect(() => {
    closeRef.current?.focus();
    const close = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', close);
    return () => window.removeEventListener('keydown', close);
  }, [onClose]);
  const date = (value?: string) =>
    value
      ? new Intl.DateTimeFormat(i18n.language, { dateStyle: 'medium', timeStyle: 'short' }).format(
          new Date(value)
        )
      : '';

  return (
    <aside className="buddy-drawer" aria-label={t('common.caret.buddy.routines')}>
      <header className="buddy-drawer-header">
        <h2>{t('common.caret.buddy.routines')}</h2>
        <button
          ref={closeRef}
          type="button"
          className="portal-icon-button"
          onClick={onClose}
          aria-label={t('common.close')}
        >
          <X size={16} />
        </button>
      </header>
      <div className="buddy-drawer-body">
        <section className="buddy-look" aria-label={t('common.caret.buddy.look')}>
          <BuddySprite
            size={52}
            color={BUDDY_COLORS[color]}
            pokeable
            title={t('common.caret.buddy.title')}
          />
          <div>
            <strong>{t('common.caret.buddy.look')}</strong>
            <div
              className="buddy-swatches"
              role="radiogroup"
              aria-label={t('common.caret.buddy.look')}
            >
              {(Object.keys(BUDDY_COLORS) as BuddyColor[]).map((name) => (
                <button
                  key={name}
                  type="button"
                  role="radio"
                  aria-checked={color === name}
                  aria-label={t(`common.caret.buddy.colors.${name}`)}
                  title={t(`common.caret.buddy.colors.${name}`)}
                  className="buddy-swatch"
                  data-color={name}
                  style={BUDDY_COLORS[name] ? { background: BUDDY_COLORS[name] } : undefined}
                  onClick={() => setColor(name)}
                />
              ))}
            </div>
          </div>
        </section>
        <p className="buddy-note">{t('common.caret.buddy.routinesHint')}</p>
        {routines.isLoading && <p role="status">{t('common.caret.buddy.loading')}</p>}
        {routines.isError && (
          <p role="alert" className="buddy-error">
            {t('common.caret.buddy.routinesFailed')}
          </p>
        )}
        {routines.data?.length === 0 && (
          <p className="buddy-empty">{t('common.caret.buddy.noRoutines')}</p>
        )}
        <ul className="buddy-routines">
          {routines.data?.map((routine) => (
            <li key={routine.id} className="buddy-routine">
              <div className="buddy-routine-text">
                <strong>{routine.name}</strong>
                <span>
                  {every(routine.everyMinutes, t)}
                  {routine.enabled && routine.nextRunAt
                    ? ` · ${t('common.caret.buddy.next', { date: date(routine.nextRunAt) })}`
                    : ''}
                  {!routine.enabled ? ` · ${t('common.caret.buddy.paused')}` : ''}
                </span>
                <p>{routine.instructions}</p>
              </div>
              <div className="buddy-routine-actions">
                <label className="buddy-switch">
                  <input
                    type="checkbox"
                    checked={routine.enabled ?? false}
                    disabled={toggle.isPending}
                    onChange={() => toggle.mutate(routine)}
                    aria-label={t('common.caret.buddy.toggleRoutine', { name: routine.name })}
                  />
                  <span />
                </label>
                <button
                  type="button"
                  className="portal-icon-button"
                  disabled={remove.isPending}
                  onClick={() => {
                    if (
                      window.confirm(t('common.caret.buddy.deleteConfirm', { name: routine.name }))
                    )
                      remove.mutate(routine);
                  }}
                  aria-label={t('common.caret.buddy.deleteRoutine', { name: routine.name })}
                >
                  <Trash2 size={15} />
                </button>
              </div>
            </li>
          ))}
        </ul>
        {(toggle.isError || remove.isError) && (
          <p role="alert" className="buddy-error">
            {t('common.caret.buddy.routineFailed')}
          </p>
        )}
        {isAdmin && settings.data && (
          <section className="buddy-section">
            <SupportAlerts settings={settings.data} scope={scope} />
          </section>
        )}
      </div>
    </aside>
  );
}
