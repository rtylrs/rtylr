import React, { useState, useMemo, useRef } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { toast } from 'sonner';
import { Send, ChevronLeft, ChevronRight, AlertTriangle } from 'lucide-react';
import { useVirtualizer } from '@tanstack/react-virtual';
import {
  useGetRoster,
  useListShifts,
  useCreateShiftAssignment,
  usePublishRoster,
} from '../../shared/hooks/queries/useShifts';
import type { RosterEntry, ShiftAssignment } from '../../shared/hooks/queries/useShifts';
import {
  SectionHeader,
  Button,
  Skeleton,
  Card,
  Modal,
  Select,
  Label,
} from '../../shared/components/ui';
import { ErrorBoundary, EmptyState } from '@voxire/ui';
import { toLocalIso } from '../../shared/lib/dateUtils';
import { addDays } from '../../shared/lib/hrDateUtils';

const DAY_KEYS = ['mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'] as const;

const SHIFT_COLORS = [
  'var(--hr-shift-purple)',
  'var(--hr-shift-green)',
  'var(--hr-shift-amber)',
  'var(--hr-shift-red)',
  'var(--hr-shift-blue)',
  'var(--hr-shift-pink)',
  'var(--hr-shift-teal)',
];

function getWeekStart(offset: number): string {
  const now = new Date();
  const dayOfWeek = now.getDay() === 0 ? 6 : now.getDay() - 1;
  const start = new Date(now);
  start.setDate(now.getDate() - dayOfWeek + offset * 7);
  return toLocalIso(start);
}

function formatWeekLabel(weekStart: string): string {
  const start = new Date(weekStart);
  const end = new Date(start);
  end.setDate(start.getDate() + 6);
  return `${start.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })} – ${end.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })}`;
}

const ROW_HEIGHT = 56;

interface AssignModalState {
  open: boolean;
  staffId: number;
  date: string;
}

const RosterBuilderInner = (): React.JSX.Element => {
  const { t } = useTranslation();
  const [weekOffset, setWeekOffset] = useState(0);
  const weekStart = useMemo(() => getWeekStart(weekOffset), [weekOffset]);
  const { data: roster, isLoading, refetch } = useGetRoster({ weekStart });
  const { data: shifts } = useListShifts();
  const assignMutation = useCreateShiftAssignment();
  const publishMutation = usePublishRoster();

  const [assignModal, setAssignModal] = useState<AssignModalState>({
    open: false,
    staffId: 0,
    date: '',
  });
  const [selectedShiftId, setSelectedShiftId] = useState<number>(0);

  const parentRef = useRef<HTMLDivElement>(null);

  const entries: RosterEntry[] = roster ?? [];

  const rowVirtualizer = useVirtualizer({
    count: entries.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 5,
  });

  const weekDates = useMemo((): string[] => {
    // addDays does the arithmetic in UTC on the ISO string. Building a Date
    // from weekStart and stepping it with a local setter lost the day the
    // clocks go forward, which put a whole roster week on the wrong dates.
    return Array.from({ length: 7 }, (_, i) => addDays(weekStart, i));
  }, [weekStart]);

  const openAssignModal = (staffId: number, date: string): void => {
    setAssignModal({ open: true, staffId, date });
    setSelectedShiftId(0);
  };

  const handleAssign = (): void => {
    if (selectedShiftId === 0) return;
    assignMutation.mutate(
      { staffId: assignModal.staffId, shiftId: selectedShiftId, date: assignModal.date },
      {
        onSuccess: () => setAssignModal({ open: false, staffId: 0, date: '' }),
        onError: () => toast.error(t('errors.assignShift')),
      }
    );
  };

  const handlePublish = (): void => {
    if (hasConflicts) {
      toast.error(t('shifts.roster.cannotPublishConflicts'));
      return;
    }
    publishMutation.mutate(weekStart, {
      onSuccess: () => toast.success(t('shifts.roster.published')),
      onError: () => toast.error(t('errors.publishRoster')),
    });
  };

  const hasConflicts = entries.some((e) => e.assignments.some((a) => a.conflict));

  if (isLoading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <Skeleton height={32} width={200} />
        <Skeleton height={400} />
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <SectionHeader
        title={t('shifts.roster.title')}
        subtitle={formatWeekLabel(weekStart)}
        right={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
            <Button
              variant="ghost"
              onClick={() => setWeekOffset((o) => o - 1)}
              aria-label={t('shifts.prevWeek')}
            >
              <ChevronLeft size={14} />
            </Button>
            <Button
              variant="ghost"
              onClick={() => setWeekOffset((o) => o + 1)}
              aria-label={t('shifts.nextWeek')}
            >
              <ChevronRight size={14} />
            </Button>
            <Button
              variant="primary"
              icon={<Send size={13} />}
              onClick={handlePublish}
              disabled={publishMutation.isPending}
            >
              {publishMutation.isPending ? t('common.loading') : t('shifts.roster.publish')}
            </Button>
          </div>
        }
      />

      {hasConflicts && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            padding: '10px 14px',
            borderRadius: 8,
            background: 'color-mix(in srgb, var(--hr-shift-red) 8%, transparent)',
            border: '1px solid color-mix(in srgb, var(--hr-shift-red) 20%, transparent)',
            color: 'var(--hr-shift-red)',
            fontSize: 12,
            fontWeight: 600,
          }}
        >
          <AlertTriangle size={14} />
          {t('shifts.roster.conflictsWarning')}
        </div>
      )}

      {entries.length === 0 ? (
        <EmptyState
          icon={AlertTriangle}
          title={t('shifts.roster.emptyTitle')}
          description={t('shifts.roster.emptyDescription')}
          action={{ label: t('common.refresh'), onClick: () => refetch() }}
        />
      ) : (
        <Card style={{ padding: 0, overflow: 'hidden' }}>
          {/* Header row */}
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: '160px repeat(7, 1fr)',
              borderBottom: '1px solid var(--hr-border)',
            }}
          >
            <div
              style={{
                padding: '10px 14px',
                fontSize: 11,
                fontWeight: 600,
                color: 'var(--hr-text-muted)',
                textTransform: 'uppercase',
              }}
            >
              {t('shifts.roster.employee')}
            </div>
            {DAY_KEYS.map((key, i) => (
              <div
                key={key}
                style={{
                  padding: '10px 8px',
                  fontSize: 11,
                  fontWeight: 600,
                  color: 'var(--hr-text-muted)',
                  textTransform: 'uppercase',
                  textAlign: 'center',
                  borderLeft: '1px solid var(--hr-border)',
                }}
              >
                {t(`timesheets.days.${key}`)}
                <div
                  style={{
                    fontSize: 10,
                    fontWeight: 500,
                    color: 'var(--hr-text-dim)',
                    marginTop: 2,
                  }}
                >
                  {new Date(weekDates[i]).getDate()}
                </div>
              </div>
            ))}
          </div>

          {/* Body with virtualisation */}
          <div
            ref={parentRef}
            style={{
              maxHeight: 20 * ROW_HEIGHT,
              overflowY: 'auto',
              position: 'relative',
            }}
          >
            <div style={{ height: rowVirtualizer.getTotalSize(), position: 'relative' }}>
              {rowVirtualizer.getVirtualItems().map((virtualRow) => {
                const entry = entries[virtualRow.index];
                return (
                  <div
                    key={entry.staffId}
                    style={{
                      position: 'absolute',
                      top: 0,
                      left: 0,
                      width: '100%',
                      height: virtualRow.size,
                      transform: `translateY(${virtualRow.start}px)`,
                      display: 'grid',
                      gridTemplateColumns: '160px repeat(7, 1fr)',
                      borderBottom: '1px solid var(--hr-border)',
                    }}
                  >
                    <div
                      style={{
                        padding: '10px 14px',
                        display: 'flex',
                        alignItems: 'center',
                        fontSize: 12,
                        fontWeight: 600,
                        color: 'var(--hr-text)',
                      }}
                    >
                      {entry.staffName}
                    </div>
                    {weekDates.map((date) => {
                      const cellAssignments = entry.assignments.filter((a) => a.date === date);
                      return (
                        <div
                          key={date}
                          onClick={() => openAssignModal(entry.staffId, date)}
                          style={{
                            padding: 4,
                            borderLeft: '1px solid var(--hr-border)',
                            cursor: 'pointer',
                            display: 'flex',
                            flexDirection: 'column',
                            gap: 2,
                            minHeight: 48,
                          }}
                          role="button"
                          tabIndex={0}
                          aria-label={t('shifts.roster.assignShift')}
                          onKeyDown={(e) => {
                            if (e.key === 'Enter' || e.key === ' ')
                              openAssignModal(entry.staffId, date);
                          }}
                        >
                          {cellAssignments.map((a) => (
                            <ShiftChip key={a.id} assignment={a} />
                          ))}
                        </div>
                      );
                    })}
                  </div>
                );
              })}
            </div>
          </div>
        </Card>
      )}

      {/* Assign shift modal */}
      <Modal
        open={assignModal.open}
        onClose={() => setAssignModal({ open: false, staffId: 0, date: '' })}
        title={t('shifts.roster.assignShift')}
        width={360}
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div>
            <Label>{t('shifts.roster.selectShift')}</Label>
            <Select
              value={selectedShiftId}
              onChange={(e) => setSelectedShiftId(Number(e.target.value))}
            >
              <option value={0}>{t('shifts.roster.choosePlaceholder')}</option>
              {(shifts ?? []).map((s) => (
                <option key={s.id} value={s.id}>
                  {s.startTime} – {s.endTime} ({s.name})
                </option>
              ))}
            </Select>
          </div>
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
            <Button
              variant="ghost"
              onClick={() => setAssignModal({ open: false, staffId: 0, date: '' })}
            >
              {t('common.cancel')}
            </Button>
            <Button
              variant="primary"
              onClick={handleAssign}
              disabled={selectedShiftId === 0 || assignMutation.isPending}
            >
              {assignMutation.isPending ? t('common.loading') : t('shifts.roster.assign')}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};

const ShiftChip = ({ assignment }: { assignment: ShiftAssignment }): React.JSX.Element => {
  const color = assignment.color || SHIFT_COLORS[assignment.shiftId % SHIFT_COLORS.length];
  return (
    <div
      style={{
        padding: '2px 6px',
        borderRadius: 4,
        fontSize: 9,
        fontWeight: 600,
        background: assignment.conflict
          ? 'color-mix(in srgb, var(--hr-shift-red) 15%, transparent)'
          : `${color}22`,
        color: assignment.conflict ? 'var(--hr-shift-red)' : color,
        border: assignment.conflict
          ? '1px solid color-mix(in srgb, var(--hr-shift-red) 30%, transparent)'
          : `1px solid ${color}33`,
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        whiteSpace: 'nowrap',
      }}
    >
      {assignment.startTime}–{assignment.endTime}
    </div>
  );
};

const RosterBuilder = (): React.JSX.Element => {
  const { t } = useTranslation();
  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.rosterBuilder'))} />
      <ErrorBoundary>
        <RosterBuilderInner />
      </ErrorBoundary>
    </>
  );
};

export default RosterBuilder;
