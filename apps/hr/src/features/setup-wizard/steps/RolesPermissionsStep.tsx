import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import StepContainer from '../components/StepContainer';

interface Permission {
  key: string;
  label: string;
  isGranted: boolean;
}

interface RoleRow {
  name: string;
  permissions: Permission[];
}

const DEFAULT_PERMISSIONS: Permission[] = [
  { key: 'hr.employees.view', label: 'View employees', isGranted: true },
  { key: 'hr.employees.create', label: 'Create employees', isGranted: false },
  { key: 'hr.employees.edit', label: 'Edit employees', isGranted: false },
  { key: 'hr.payroll.view', label: 'View payroll', isGranted: false },
  { key: 'hr.payroll.run', label: 'Run payroll', isGranted: false },
  { key: 'hr.shifts.manage', label: 'Manage shifts', isGranted: false },
  { key: 'hr.time_off.approve', label: 'Approve time off', isGranted: false },
  { key: 'hr.reports.view', label: 'View reports', isGranted: false },
];

const DEFAULT_ROLES: RoleRow[] = [
  { name: 'HR Admin', permissions: DEFAULT_PERMISSIONS.map((p) => ({ ...p, isGranted: true })) },
  {
    name: 'HR Manager',
    permissions: DEFAULT_PERMISSIONS.map((p) => ({
      ...p,
      isGranted: [
        'hr.employees.view',
        'hr.employees.create',
        'hr.employees.edit',
        'hr.shifts.manage',
        'hr.time_off.approve',
        'hr.reports.view',
      ].includes(p.key),
    })),
  },
  {
    name: 'Payroll Admin',
    permissions: DEFAULT_PERMISSIONS.map((p) => ({
      ...p,
      isGranted: ['hr.employees.view', 'hr.payroll.view', 'hr.payroll.run'].includes(p.key),
    })),
  },
  {
    name: 'Shift Manager',
    permissions: DEFAULT_PERMISSIONS.map((p) => ({
      ...p,
      isGranted: ['hr.employees.view', 'hr.shifts.manage'].includes(p.key),
    })),
  },
  {
    name: 'Employee',
    permissions: DEFAULT_PERMISSIONS.map((p) => ({
      ...p,
      isGranted: p.key === 'hr.employees.view',
    })),
  },
];

interface RolesPermissionsStepProps {
  onNext: (roles: RoleRow[]) => void;
  onBack: () => void;
  isSubmitting?: boolean;
}

const RolesPermissionsStep = ({
  onNext,
  onBack,
  isSubmitting,
}: RolesPermissionsStepProps): React.JSX.Element => {
  const { t } = useTranslation();
  const [roles, setRoles] = useState<RoleRow[]>(DEFAULT_ROLES);
  const [expandedRole, setExpandedRole] = useState<number | null>(0);

  const updateRoleName = (index: number, name: string): void => {
    setRoles((prev) => prev.map((r, i) => (i === index ? { ...r, name } : r)));
  };

  const togglePermission = (roleIndex: number, permKey: string): void => {
    setRoles((prev) =>
      prev.map((r, i) =>
        i === roleIndex
          ? {
              ...r,
              permissions: r.permissions.map((p) =>
                p.key === permKey ? { ...p, isGranted: !p.isGranted } : p
              ),
            }
          : r
      )
    );
  };

  const handleSubmit = (): void => {
    onNext(roles);
  };

  return (
    <StepContainer
      title={t('setup.rolesPermissions.title')}
      description={t('setup.rolesPermissions.description')}
      onBack={onBack}
      onNext={handleSubmit}
      isSubmitting={isSubmitting}
    >
      <div className="flex flex-col gap-3">
        {roles.map((role, roleIndex) => (
          <div
            key={roleIndex}
            className="rounded-lg overflow-hidden"
            style={{ border: '1px solid var(--hr-border)' }}
          >
            <button
              type="button"
              onClick={() => setExpandedRole(expandedRole === roleIndex ? null : roleIndex)}
              className="flex w-full items-center justify-between px-4 py-3 text-left"
              style={{ background: 'var(--hr-hover)' }}
            >
              <input
                type="text"
                value={role.name}
                onChange={(e) => updateRoleName(roleIndex, e.target.value)}
                onClick={(e) => e.stopPropagation()}
                className="bg-transparent text-sm font-medium outline-none"
                style={{ color: 'var(--hr-text)' }}
                aria-label={t('setup.rolesPermissions.roleName')}
              />
              <span className="text-xs" style={{ color: 'var(--hr-text-dim)' }}>
                {role.permissions.filter((p) => p.isGranted).length}/{role.permissions.length}
              </span>
            </button>
            {expandedRole === roleIndex && (
              <div className="px-4 py-3" style={{ background: 'var(--hr-card)' }}>
                <div className="flex flex-col gap-2">
                  {role.permissions.map((perm) => (
                    <label key={perm.key} className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={perm.isGranted}
                        onChange={() => togglePermission(roleIndex, perm.key)}
                        className="h-3.5 w-3.5 rounded"
                      />
                      <span className="text-xs" style={{ color: 'var(--hr-text-muted)' }}>
                        {perm.label}
                      </span>
                    </label>
                  ))}
                </div>
              </div>
            )}
          </div>
        ))}
      </div>
    </StepContainer>
  );
};

export default RolesPermissionsStep;
