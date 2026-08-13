import React, { useEffect, useRef, useState } from 'react';
import { Plus, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Alert, Avatar } from '@voxire/ui';
import {
  type RoundingMinutes,
  type WeekStart,
  useTimeSettingsStore,
} from '../../../shared/stores/timeSettingsStore';
import { useFlowTeam } from '../../../shared/hooks/queries/useFlowTeam';
import { initialsFromName } from '../../../shared/lib/memberUtils';
import { formatDurationClock } from '../../../shared/lib/durationUtils';
import GateNotice from '../../../shared/components/GateNotice';
import FlowSwitch from '../../../shared/components/FlowSwitch';

const ROUNDING_OPTIONS: RoundingMinutes[] = [0, 5, 10, 15, 30];
const UNDO_WINDOW_MS = 2000;

const labelClass = 'mb-1 block text-[10px] font-semibold uppercase tracking-wider';
const inputClass =
  'flow-focus min-h-[44px] w-full rounded px-2 py-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]';

function inputStyle(): React.CSSProperties {
  return {
    background: 'var(--flow-input-bg)',
    border: '1px solid var(--flow-border)',
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
  const { data: team = [] } = useFlowTeam();

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
      <Alert
        variant="warning"
        data-c="alert"
        className="border-[var(--flow-status-pending-border)] bg-[var(--flow-status-pending-bg)] text-[var(--flow-text)]"
      >
        <p className="text-[13px] leading-snug">{t('flow.settings.time.localOnlyWarning')}</p>
      </Alert>

      <section>
        <h2
          className="mb-3 text-[11px] font-semibold uppercase tracking-wider"
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

          <div>
            <label
              htmlFor="time-weekly-target"
              className={labelClass}
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.settings.time.weeklyTarget')}
            </label>
            <input
              id="time-weekly-target"
              type="number"
              min={1}
              max={168}
              value={Math.round(settings.weeklyTargetMinutes / 60)}
              onChange={(e) => {
                const hours = Number(e.target.value);
                if (!Number.isFinite(hours) || hours <= 0) return;
                settings.setWeeklyTargetMinutes(hours * 60);
              }}
              className={inputClass}
              style={inputStyle()}
            />
            <p className="mt-1 text-[10px]" style={{ color: 'var(--flow-text-muted)' }}>
              {formatDurationClock(settings.weeklyTargetMinutes)}
            </p>
          </div>

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
                if (hours * 60 < settings.weeklyTargetMinutes) {
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
              <p className="mt-1 text-[10px]" style={{ color: 'var(--flow-status-rejected)' }}>
                {t('flow.settings.time.overtimeBelowTarget')}
              </p>
            )}
          </div>
        </div>
      </section>

      <section>
        <h2
          className="mb-3 text-[11px] font-semibold uppercase tracking-wider"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.settings.time.approval')}
        </h2>

        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0 flex-1">
            <p className="text-[13px] leading-snug" style={{ color: 'var(--flow-text)' }}>
              {t('flow.settings.time.approvalLocksExplained')}
            </p>
            {/* Stated before the switch is touched: both directions are
             * irreversible for entries that already exist. */}
            <p
              className="mt-1 text-[11px] leading-snug"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.settings.time.lockConsequence')}
            </p>
            {lockUndo !== null && (
              <button
                type="button"
                onClick={undoLock}
                className="flow-focus mt-2 min-h-[44px] rounded px-2 text-[11px] font-semibold md:min-h-0 md:py-1"
                style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
              >
                {t('flow.settings.time.undo')}
              </button>
            )}
          </div>
          <FlowSwitch
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
          <p className="mt-1 text-[11px] leading-snug" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.settings.time.autoLockExplained', { count: settings.autoLockDays })}
          </p>
        </div>
      </section>

      <section>
        <h2
          className="mb-3 text-[11px] font-semibold uppercase tracking-wider"
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
                className="flow-focus min-h-[44px] flex-1 rounded bg-transparent px-2 text-[16px] md:min-h-0 md:py-1 md:text-[13px]"
                style={{ color: 'var(--flow-text)' }}
              />
              <FlowSwitch
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
                className="flow-focus flex h-11 w-11 shrink-0 items-center justify-center rounded disabled:opacity-30 md:h-8 md:w-8"
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
            className="flow-focus inline-flex min-h-[44px] items-center justify-center gap-1.5 rounded px-3 text-[13px] font-semibold md:min-h-0 md:py-2"
            style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
          >
            <Plus size={14} aria-hidden="true" />
            {t('flow.settings.time.addCategory')}
          </button>
        </div>
      </section>

      <section>
        <h2
          className="mb-3 text-[11px] font-semibold uppercase tracking-wider"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.settings.time.billableRates')}
        </h2>
        <p className="mb-3 text-[11px] leading-snug" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.settings.time.ratesOnePerPerson')}
        </p>

        <ul className="space-y-1.5">
          {team.map((member) => {
            const rate = settings.ratesByUserId[member.id];
            const hasRate = typeof rate === 'number';
            return (
              <li
                key={member.id}
                className="flex flex-col gap-2 rounded-md p-2 sm:flex-row sm:items-center sm:gap-3"
                style={{ background: 'var(--flow-hover)' }}
              >
                <div className="flex min-w-0 flex-1 items-center gap-2">
                  <Avatar size="sm" initials={initialsFromName(member.name)} aria-hidden="true" />
                  <span className="truncate text-[13px]" style={{ color: 'var(--flow-text)' }}>
                    {member.name}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <input
                    type="number"
                    min={0}
                    step="0.01"
                    aria-label={t('flow.settings.time.rateFor', { name: member.name })}
                    value={hasRate ? rate : ''}
                    onChange={(e) => {
                      const raw = e.target.value.trim();
                      // Empty clears back to unset. A zero rate silently bills
                      // nothing, so it is never the fallback for a blank field.
                      if (raw === '') {
                        settings.setRate(member.id, null);
                        return;
                      }
                      const next = Number(raw);
                      if (!Number.isFinite(next) || next < 0) return;
                      settings.setRate(member.id, next);
                    }}
                    className="flow-focus min-h-[44px] w-24 rounded px-2 text-[16px] md:min-h-0 md:py-1 md:text-[13px]"
                    style={inputStyle()}
                  />
                  {!hasRate && (
                    <span className="text-[11px]" style={{ color: 'var(--flow-text-muted)' }}>
                      {t('flow.settings.time.rateNotSet')}
                    </span>
                  )}
                </div>
                {/* The per-project override column ships empty on purpose: rates
                 * arrive per project as a value in this cell, not a rewrite. */}
                {/* Reserves the per-project override column without drawing a
                  * value in it. Nothing is there yet, and a placeholder glyph
                  * would read as a value of "none" rather than "not built". */}
                <span className="hidden sm:block sm:w-28" aria-hidden="true" />
              </li>
            );
          })}
        </ul>
      </section>
    </div>
  );
};

export default TimeSection;
