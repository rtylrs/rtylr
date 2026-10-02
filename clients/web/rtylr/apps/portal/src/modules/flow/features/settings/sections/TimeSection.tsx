import React, { useEffect, useRef, useState } from 'react';
import { Plus, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Alert, Switch } from '@voxire/ui';
import {
  type RoundingMinutes,
  type WeekStart,
  getWeeklyTargetMinutes,
  useTimeSettingsStore,
} from '../../../shared/stores/timeSettingsStore';
import { DAY_KEYS, type DayKey } from '../../../shared/lib/dateUtils';
import { formatDurationClock } from '../../../shared/lib/durationUtils';
import GateNotice from '../../../shared/components/GateNotice';
import WorkTypesSection from './WorkTypesSection';

const ROUNDING_OPTIONS: RoundingMinutes[] = [0, 5, 10, 15, 30];
const UNDO_WINDOW_MS = 2000;
// i18n keys for each day already exist as flow.settings.time.<day> (used by
// the week-start dropdown), so daily targets reuse the same keys rather than
// duplicating the seven day names under a new namespace.
const DAY_LABEL_KEYS: Record<DayKey, string> = {
  sun: 'sunday',
  mon: 'monday',
  tue: 'tuesday',
  wed: 'wednesday',
  thu: 'thursday',
  fri: 'friday',
  sat: 'saturday',
};

const labelClass = 'mb-1 block text-flow-meta font-semibold uppercase tracking-wider';
const inputClass =
  'flow-focus min-h-[44px] w-full rounded-sm px-2 py-2 text-flow-title md:min-h-0 md:py-1.5 md:text-flow-body';

function inputStyle(): React.CSSProperties {
  return {
    background: 'var(--flow-input-bg)',
    border: '1px solid var(--flow-border-control)',
    color: 'var(--flow-text)',
  };
}

/**
 * Settings → Time. A new section in the existing settings rail, not a new screen.
 *
 * Every value here is a workspace value with a shipped default, because a law
 * firm and a restaurant need different defaults rather than different code.
 *
 * There is no workspace time-settings endpoint yet, so these persist per-browser
 * through useTimeSettingsStore and the pane says so at the top. Categories and
 * rates additionally have no server field to attach to - stated inline rather
 * than implied by a control that appears to work.
 */
const TimeSection = (): React.JSX.Element => {
  const { t } = useTranslation();
  const settings = useTimeSettingsStore();

  const [overtimeError, setOvertimeError] = useState(false);
  const [newCategory, setNewCategory] = useState('');
  const [lockUndo, setLockUndo] = useState<boolean | null>(null);
  const undoTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(
    () => () => {
      if (undoTimer.current) clearTimeout(undoTimer.current);
    },
    []
  );

  const flipLock = (next: boolean): void => {
    const previous = settings.approvalLocksEntry;
    settings.setApprovalLocksEntry(next);
    setLockUndo(previous);
    if (undoTimer.current) clearTimeout(undoTimer.current);
    undoTimer.current = setTimeout(() => setLockUndo(null), UNDO_WINDOW_MS);
  };

  const undoLock = (): void => {
    if (lockUndo === null) return;
    settings.setApprovalLocksEntry(lockUndo);
    setLockUndo(null);
    if (undoTimer.current) clearTimeout(undoTimer.current);
  };

  const addCategory = (): void => {
    settings.addCategory(newCategory, true);
    setNewCategory('');
  };

  return (
    <div className="space-y-8">
      {/* The honest header. Without it this pane reads as a workspace setting
       * that quietly is not one. */}
      <Alert variant="warning" data-c="alert">
        <p className="text-flow-body leading-snug">{t('flow.settings.time.localOnlyWarning')}</p>
      </Alert>

      <section>
        <h2
          className="mb-3 text-flow-meta font-semibold uppercase tracking-wider"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.settings.time.weekAndTargets')}
        </h2>

        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <label
              htmlFor="time-week-start"
              className={labelClass}
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.settings.time.weekStart')}
            </label>
            <select
              id="time-week-start"
              value={settings.weekStart}
              onChange={(e) => settings.setWeekStart(e.target.value as WeekStart)}
              className={inputClass}
              style={inputStyle()}
            >
              <option value="monday">{t('flow.settings.time.monday')}</option>
              <option value="sunday">{t('flow.settings.time.sunday')}</option>
            </select>
          </div>

          <div>
            <label
              htmlFor="time-rounding"
              className={labelClass}
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.settings.time.rounding')}
            </label>
            <select
              id="time-rounding"
              value={settings.roundingMinutes}
              onChange={(e) =>
                settings.setRoundingMinutes(Number(e.target.value) as RoundingMinutes)
              }
              className={inputClass}
              style={inputStyle()}
            >
              {ROUNDING_OPTIONS.map((m) => (
                <option key={m} value={m}>
                  {m === 0
                    ? t('flow.settings.time.roundingOff')
                    : t('flow.settings.time.roundingNearest', { minutes: m })}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Daily targets: fully independent per day, not a weekly figure split
         * evenly - a 4-day week, a short Friday, or a non-working Sunday are
         * real shapes a business has, not deviations from a uniform one. */}
        <div className="mt-6">
          <label
            className={labelClass}
            style={{ color: 'var(--flow-text-muted)' }}
            id="time-daily-targets-label"
          >
            {t('flow.settings.time.dailyTargets')}
          </label>
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4 lg:grid-cols-7">
            {DAY_KEYS.map((day) => (
              <div key={day}>
                <label
                  htmlFor={`time-daily-target-${day}`}
                  className="mb-1 block text-flow-meta font-medium"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  {t(`flow.settings.time.${DAY_LABEL_KEYS[day]}`)}
                </label>
                <input
                  id={`time-daily-target-${day}`}
                  type="number"
                  min={0}
                  max={24}
                  step={0.5}
                  value={settings.dailyTargetMinutes[day] / 60}
                  onChange={(e) => {
                    const hours = Number(e.target.value);
                    if (!Number.isFinite(hours) || hours < 0) return;
                    settings.setDailyTargetMinutes(day, Math.round(hours * 60));
                  }}
                  className={inputClass}
                  style={inputStyle()}
                />
              </div>
            ))}
          </div>
          <p className="mt-2 text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.settings.time.weeklyTargetSummary', {
              total: formatDurationClock(getWeeklyTargetMinutes(settings.dailyTargetMinutes)),
            })}
          </p>
        </div>

        <div className="mt-6 grid gap-4 sm:grid-cols-2">
          <div>
            <label
              htmlFor="time-overtime"
              className={labelClass}
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.settings.time.overtimeThreshold')}
            </label>
            <input
              id="time-overtime"
              type="number"
              min={1}
              max={168}
              value={Math.round(settings.overtimeThresholdMinutes / 60)}
              aria-invalid={overtimeError || undefined}
              onChange={(e) => {
                const hours = Number(e.target.value);
                if (!Number.isFinite(hours) || hours <= 0) return;
                // A threshold under the target would warn on a normal week, so
                // the field reverts and says why rather than accepting it.
                const weeklyTargetMinutes = getWeeklyTargetMinutes(settings.dailyTargetMinutes);
                if (hours * 60 < weeklyTargetMinutes) {
                  setOvertimeError(true);
                  return;
                }
                setOvertimeError(false);
                settings.setOvertimeThresholdMinutes(hours * 60);
              }}
              className={inputClass}
              style={inputStyle()}
            />
            {overtimeError && (
              <p className="mt-1 text-flow-meta" style={{ color: 'var(--flow-status-rejected)' }}>
                {t('flow.settings.time.overtimeBelowTarget')}
              </p>
            )}
          </div>
        </div>
      </section>

      <section>
        <h2
          className="mb-3 text-flow-meta font-semibold uppercase tracking-wider"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.settings.time.approval')}
        </h2>

        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0 flex-1">
            <p className="text-flow-body leading-snug" style={{ color: 'var(--flow-text)' }}>
              {t('flow.settings.time.approvalLocksExplained')}
            </p>
            {/* Stated before the switch is touched: both directions are
             * irreversible for entries that already exist. */}
            <p
              className="mt-1 text-flow-meta leading-snug"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.settings.time.lockConsequence')}
            </p>
            {lockUndo !== null && (
              <button
                type="button"
                onClick={undoLock}
                className="flow-focus mt-2 min-h-[44px] rounded-sm px-2 text-flow-meta font-semibold md:min-h-0 md:py-1"
                style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
              >
                {t('flow.settings.time.undo')}
              </button>
            )}
          </div>
          <Switch
            checked={settings.approvalLocksEntry}
            onCheckedChange={flipLock}
            aria-label={t('flow.settings.time.approvalLocks')}
            className="mt-0.5 shrink-0"
          />
        </div>

        <div className="mt-4">
          <label
            htmlFor="time-autolock"
            className={labelClass}
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.settings.time.autoLockDays')}
          </label>
          <input
            id="time-autolock"
            type="number"
            min={1}
            max={90}
            value={settings.autoLockDays}
            onChange={(e) => {
              const days = Number(e.target.value);
              if (!Number.isFinite(days) || days <= 0) return;
              settings.setAutoLockDays(days);
            }}
            className={`${inputClass} sm:max-w-[160px]`}
            style={inputStyle()}
          />
          <p
            className="mt-1 text-flow-meta leading-snug"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.settings.time.autoLockExplained', { count: settings.autoLockDays })}
          </p>
        </div>
      </section>

      <section>
        <h2
          className="mb-3 text-flow-meta font-semibold uppercase tracking-wider"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.settings.time.attendance')}
        </h2>

        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0 flex-1">
            <p className="text-flow-body leading-snug" style={{ color: 'var(--flow-text)' }}>
              {t('flow.settings.time.trackAttendanceExplained')}
            </p>
          </div>
          <Switch
            checked={settings.trackAttendance}
            onCheckedChange={settings.setTrackAttendance}
            aria-label={t('flow.settings.time.trackAttendance')}
            className="mt-0.5 shrink-0"
          />
        </div>
      </section>

      <WorkTypesSection />

      <section>
        <h2
          className="mb-3 text-flow-meta font-semibold uppercase tracking-wider"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.settings.time.categories')}
        </h2>

        <GateNotice className="mb-3" reason={t('flow.settings.time.categoriesNotWired')} />

        <ul className="space-y-1.5">
          {settings.categories.map((category) => (
            <li
              key={category.id}
              className="flex flex-col gap-2 rounded-md p-2 sm:flex-row sm:items-center sm:gap-3"
              style={{ background: 'var(--flow-hover)' }}
            >
              <input
                aria-label={t('flow.settings.time.categoryName')}
                value={category.name}
                onChange={(e) => settings.renameCategory(category.id, e.target.value)}
                className="flow-focus min-h-[44px] flex-1 rounded-sm bg-transparent px-2 text-flow-title md:min-h-0 md:py-1 md:text-flow-body"
                style={{ color: 'var(--flow-text)' }}
              />
              <Switch
                checked={category.billable}
                onCheckedChange={(next) => settings.setCategoryBillable(category.id, next)}
                aria-label={t('flow.settings.time.categoryBillable')}
                className="shrink-0"
              />
              <button
                type="button"
                onClick={() => settings.deleteCategory(category.id)}
                disabled={settings.categories.length === 1}
                aria-label={t('flow.settings.time.removeCategoryLabel', { name: category.name })}
                className="flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm disabled:opacity-30 md:h-8 md:w-8"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                <X size={14} aria-hidden="true" />
              </button>
            </li>
          ))}
        </ul>

        <div className="mt-3 flex flex-col gap-2 sm:flex-row">
          <input
            aria-label={t('flow.settings.time.newCategoryName')}
            value={newCategory}
            onChange={(e) => setNewCategory(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault();
                addCategory();
              }
            }}
            placeholder={t('flow.settings.time.newCategoryPlaceholder')}
            className={`${inputClass} sm:flex-1`}
            style={inputStyle()}
          />
          <button
            type="button"
            onClick={addCategory}
            className="flow-focus inline-flex min-h-[44px] items-center justify-center gap-1.5 rounded-sm px-3 text-flow-body font-semibold md:min-h-0 md:py-2"
            style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
          >
            <Plus size={14} aria-hidden="true" />
            {t('flow.settings.time.addCategory')}
          </button>
        </div>
      </section>
    </div>
  );
};

export default TimeSection;
