import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import { useStaffReportingLines, useStaffRecord, useSetStaffManager } from './useStaff';

/**
 * Reporting lines: who each staff member reports to.
 *
 * apps/hr had no staff edit screen at all - useCreateStaff and useUpdateStaff
 * existed and no component imported them - so F-B's manager field had nowhere to
 * be set. This is the narrowest screen that fixes that: it edits exactly one
 * field and nothing else.
 *
 * It writes through the GENERATED client rather than hrApi.updateStaff, which
 * puts to /hr/staff with no id while the contract declares PUT /hr/staff/{id}.
 * That mismatch is why the legacy path 404s.
 *
 * Setting a manager here widens what that person can see in Flow: the
 * dept-manager visibility proxy unions this reporting line with the
 * department-manager relationship. The screen says so, because that consequence
 * is not obvious from an HR form.
 */
const ReportingLines: React.FC = () => {
  const { t } = useTranslation();
  const { data: staff = [], isLoading } = useStaffReportingLines();
  const [editingId, setEditingId] = useState<number | null>(null);
  const { data: record } = useStaffRecord(editingId);
  const setManager = useSetStaffManager();

  const save = (managerStaffId: number): void => {
    if (!record) return;
    setManager.mutate(
      { staff: record, managerStaffId },
      {
        onSuccess: () => {
          toast.success(t('staff.reportingLines.saved'));
          setEditingId(null);
        },
      }
    );
  };

  const nameFor = (id: number): string =>
    staff.find((s) => s.id === id)?.name ?? t('staff.reportingLines.none');

  return (
    <div className="h-full overflow-y-auto">
      <HelmetContainer title={t('staff.reportingLines.title')} />
      <div className="mx-auto max-w-2xl px-4 py-6 sm:px-6">
        <h1 className="text-xl font-bold" style={{ color: 'var(--hr-text)' }}>
          {t('staff.reportingLines.title')}
        </h1>
        <p className="mt-1 mb-6 text-sm" style={{ color: 'var(--hr-text-muted)' }}>
          {t('staff.reportingLines.flowNotice')}
        </p>

        {isLoading ? (
          <p className="text-sm" style={{ color: 'var(--hr-text-muted)' }}>
            {t('common.loading')}
          </p>
        ) : staff.length === 0 ? (
          <p className="text-sm" style={{ color: 'var(--hr-text-muted)' }}>
            {t('staff.reportingLines.empty')}
          </p>
        ) : (
          <ul className="space-y-2">
            {staff.map((person) => (
              <li
                key={person.id}
                className="flex items-center justify-between gap-3 rounded-lg border p-3"
                style={{ borderColor: 'var(--hr-border)' }}
              >
                <div className="min-w-0">
                  <p className="text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
                    {person.name}
                  </p>
                  <p className="text-xs" style={{ color: 'var(--hr-text-muted)' }}>
                    {t('staff.reportingLines.reportsTo', {
                      name: nameFor(person.managerStaffId),
                    })}
                  </p>
                </div>

                {editingId === person.id ? (
                  <select
                    autoFocus
                    defaultValue={String(person.managerStaffId)}
                    onChange={(e) => save(Number(e.target.value))}
                    disabled={setManager.isPending || !record}
                    aria-label={t('staff.reportingLines.selectLabel', { name: person.name })}
                    className="h-11 rounded-lg border px-2 text-base"
                    style={{ borderColor: 'var(--hr-border)', color: 'var(--hr-text)' }}
                  >
                    <option value={0}>{t('staff.reportingLines.none')}</option>
                    {staff
                      // Nobody reports to themselves. The resolver also excludes
                      // it, so this keeps the UI from offering an option the
                      // server would ignore.
                      .filter((candidate) => candidate.id !== person.id)
                      .map((candidate) => (
                        <option key={candidate.id} value={candidate.id}>
                          {candidate.name}
                        </option>
                      ))}
                  </select>
                ) : (
                  <button
                    type="button"
                    onClick={() => setEditingId(person.id)}
                    className="h-11 min-w-[44px] rounded-lg px-3 text-sm font-medium"
                    style={{ color: 'var(--accent-fg)' }}
                    aria-label={t('staff.reportingLines.changeLabel', { name: person.name })}
                  >
                    {t('staff.reportingLines.change')}
                  </button>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
};

export default ReportingLines;
