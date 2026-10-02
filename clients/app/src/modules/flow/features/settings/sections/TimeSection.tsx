import React, { useState } from 'react';
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
import WorkTypesSection from './WorkTypesSection';

const ROUNDING_OPTIONS: RoundingMinutes[] = [0, 5, 10, 15, 30];
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

const TimeSection = (): React.JSX.Element => {
  const { t } = useTranslation();
  const settings = useTimeSettingsStore();

  const [overtimeError, setOvertimeError] = useState(false);
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
    </div>
  );
};

export default TimeSection;
