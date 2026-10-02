import React, { useRef, useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { useVirtualizer } from '@tanstack/react-virtual';
import { Plus, Search, Users } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { QueryError } from '../../shared/components/ui';
import { useListEmployees } from '../../shared/hooks/queries/useEmployees';
import type { EmployeePayload } from '@voxire/voxie/hr/employee/v1/types';

const PAGE_SIZE = 50;

const Employees = (): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const navigate = useNavigate();
  const [offset, setOffset] = useState(0);
  const [searchTerm, setSearchTerm] = useState('');
  const [departmentFilter, setDepartmentFilter] = useState<number | undefined>();
  const [statusFilter, setStatusFilter] = useState<string | undefined>();

  const { data, isLoading, isError, isFetching, refetch } = useListEmployees({
    limit: PAGE_SIZE,
    offset,
    departmentId: departmentFilter,
    status: statusFilter,
  });

  const employees = data?.employees ?? [];
  const total = data?.total ?? 0;

  const parentRef = useRef<HTMLDivElement>(null);

  const filteredEmployees = searchTerm
    ? employees.filter((e) =>
        Object.values(e)
          .filter((value): value is string => typeof value === 'string')
          .join(' ')
          .toLowerCase()
          .includes(searchTerm.toLowerCase())
      )
    : employees;

  // Count must track the filtered list, otherwise searching leaves the
  // virtualizer generating indices past filteredEmployees.length and the
  // row lookup returns undefined.
  const virtualizer = useVirtualizer({
    count: filteredEmployees.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 56,
    overscan: 10,
  });

  const inputStyle: React.CSSProperties = {
    padding: '8px 12px',
    borderRadius: '8px',
    border: '1px solid var(--hr-border)',
    background: 'var(--hr-input-bg)',
    color: 'var(--hr-text)',
    fontSize: '13px',
    outline: 'none',
  };

  return (
    <>
      <HelmetContainer title={pageTitle(t('hrNav.employees'))} />
      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <h1 className="text-lg font-semibold" style={{ color: 'var(--hr-text)' }}>
            {t('hrEmployees.directory')}
          </h1>
          <button
            type="button"
            onClick={() => navigate({ to: '/hr/employees/new' })}
            className="flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium"
            style={{ background: 'var(--accent)', color: 'var(--hr-on-accent)' }}
          >
            <Plus size={14} />
            {t('hrEmployees.addNew')}
          </button>
        </div>

        {isError ? (
          <QueryError onRetry={() => refetch()} isRetrying={isFetching} />
        ) : (
          <>
            {/* Filters */}
            <div className="flex flex-wrap items-center gap-2">
              <div className="relative flex-1" style={{ minWidth: 200 }}>
                <Search
                  size={14}
                  className="absolute left-3 top-1/2 -translate-y-1/2"
                  style={{ color: 'var(--hr-text-dim)' }}
                />
                <input
                  type="text"
                  placeholder={t('hrEmployees.search')}
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  style={{ ...inputStyle, paddingLeft: '32px', width: '100%' }}
                />
              </div>
              <select
                value={statusFilter ?? ''}
                onChange={(e) => setStatusFilter(e.target.value || undefined)}
                style={inputStyle}
                aria-label={t('hrEmployees.filterByStatus')}
              >
                <option value="">{t('hrEmployees.filterByStatus')}</option>
                <option value="active">{t('hrEmployees.statuses.active')}</option>
                <option value="probation">{t('hrEmployees.statuses.probation')}</option>
                <option value="inactive">{t('hrEmployees.statuses.inactive')}</option>
                <option value="terminated">{t('hrEmployees.statuses.terminated')}</option>
              </select>
            </div>

            {/* Virtualised table */}
            {isLoading ? (
              <div className="flex justify-center py-12">
                <div
                  className="h-6 w-6 rounded-full border-2 animate-spin"
                  style={{ borderColor: 'var(--hr-border)', borderTopColor: 'var(--accent)' }}
                />
              </div>
            ) : filteredEmployees.length === 0 ? (
              <EmptyState
                icon={Users}
                title={t('hrEmployees.noEmployees')}
                action={
                  searchTerm || departmentFilter !== undefined || statusFilter !== undefined
                    ? {
                        label: t('hrCommon.clearFilters'),
                        onClick: () => {
                          setSearchTerm('');
                          setDepartmentFilter(undefined);
                          setStatusFilter(undefined);
                        },
                      }
                    : {
                        label: t('hrEmployees.addNew'),
                        onClick: () => navigate({ to: '/hr/employees/new' }),
                      }
                }
              />
            ) : (
              <>
                <div
                  className="rounded-lg overflow-hidden"
                  style={{ border: '1px solid var(--hr-border)' }}
                >
                  {/* Header */}
                  <div
                    className="grid grid-cols-4 gap-4 px-4 py-2 text-xs font-medium"
                    style={{ background: 'var(--hr-hover)', color: 'var(--hr-text-muted)' }}
                  >
                    <span>{t('hrEmployees.jobTitle')}</span>
                    <span>{t('hrEmployees.contractType')}</span>
                    <span>{t('hrEmployees.hireDate')}</span>
                    <span>{t('hrCommon.status')}</span>
                  </div>

                  {/* Virtualised rows */}
                  <div
                    ref={parentRef}
                    className="overflow-auto"
                    style={{ maxHeight: 'calc(100vh - 280px)' }}
                  >
                    <div
                      style={{
                        height: `${virtualizer.getTotalSize()}px`,
                        width: '100%',
                        position: 'relative',
                      }}
                    >
                      {virtualizer.getVirtualItems().map((virtualRow) => {
                        const emp: EmployeePayload | undefined =
                          filteredEmployees[virtualRow.index];
                        if (!emp) return null;
                        return (
                          <div
                            key={virtualRow.key}
                            data-index={virtualRow.index}
                            ref={virtualizer.measureElement}
                            className="grid grid-cols-4 gap-4 px-4 py-3 cursor-pointer transition-colors"
                            style={{
                              position: 'absolute',
                              top: 0,
                              left: 0,
                              width: '100%',
                              transform: `translateY(${virtualRow.start}px)`,
                              borderTop: '1px solid var(--hr-border)',
                            }}
                            onClick={() =>
                              navigate({ to: '/hr/employees/$id', params: { id: String(emp.id) } })
                            }
                            onMouseEnter={(e) => {
                              (e.currentTarget as HTMLDivElement).style.background =
                                'var(--hr-hover)';
                            }}
                            onMouseLeave={(e) => {
                              (e.currentTarget as HTMLDivElement).style.background = 'transparent';
                            }}
                          >
                            <span className="truncate text-sm" style={{ color: 'var(--hr-text)' }}>
                              {emp.jobTitle}
                            </span>
                            <span className="text-sm" style={{ color: 'var(--hr-text-muted)' }}>
                              {t(`hrEmployees.contractTypes.${emp.contractType}` as const)}
                            </span>
                            <span className="text-sm" style={{ color: 'var(--hr-text-muted)' }}>
                              {emp.hireDate}
                            </span>
                            <span>
                              <span
                                className="rounded-full px-2 py-0.5 text-xs font-medium"
                                style={{
                                  background:
                                    emp.status === 'active'
                                      ? 'rgba(34,197,94,0.15)'
                                      : 'var(--hr-hover)',
                                  color:
                                    emp.status === 'active' ? '#22c55e' : 'var(--hr-text-muted)',
                                }}
                              >
                                {t(`hrEmployees.statuses.${emp.status}` as const)}
                              </span>
                            </span>
                          </div>
                        );
                      })}
                    </div>
                  </div>
                </div>

                {/* Pagination */}
                {total > PAGE_SIZE && (
                  <div className="flex items-center justify-center gap-4 py-2">
                    <button
                      type="button"
                      disabled={offset === 0}
                      onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
                      className="rounded px-3 py-1 text-sm disabled:opacity-40"
                      style={{ background: 'var(--hr-hover)', color: 'var(--hr-text-muted)' }}
                    >
                      {t('hrCommon.back')}
                    </button>
                    <span className="text-xs" style={{ color: 'var(--hr-text-dim)' }}>
                      {offset + 1}–{Math.min(offset + PAGE_SIZE, total)} / {total}
                    </span>
                    <button
                      type="button"
                      disabled={offset + PAGE_SIZE >= total}
                      onClick={() => setOffset(offset + PAGE_SIZE)}
                      className="rounded px-3 py-1 text-sm disabled:opacity-40"
                      style={{ background: 'var(--hr-hover)', color: 'var(--hr-text-muted)' }}
                    >
                      {t('hrCommon.next')}
                    </button>
                  </div>
                )}
              </>
            )}
          </>
        )}
      </div>
    </>
  );
};

export default Employees;
