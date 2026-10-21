import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../../shared/testing/i18n';
import enCommon from '../../../locales/en/hrCommon.json';
import enSetup from '../../../locales/en/hrSetup.json';
import RolesPermissionsStep from './RolesPermissionsStep';

// This app's i18n bootstrap (`shared/services/i18next.ts`) loads every
// namespace file into a single 'translation' namespace, keyed by filename, so
// `t('hrSetup.rolesPermissions.title')` resolves as a dot-path within that one
// namespace. Reproduce that here with the real, shipped English copy so the
// assertions below check actual production strings, not raw i18n keys.
beforeEach(async () => {
  await initTestI18n();
});

describe('RolesPermissionsStep', () => {
  it('shows the five real default role names, not the old fake taxonomy', () => {
    render(<RolesPermissionsStep onNext={vi.fn()} onBack={vi.fn()} />);

    expect(screen.getByText('Employee')).toBeInTheDocument();
    expect(screen.getByText('Manager')).toBeInTheDocument();
    expect(screen.getByText('HR Admin')).toBeInTheDocument();
    expect(screen.getByText('Finance/Payroll Manager')).toBeInTheDocument();
    expect(screen.getByText('Executive')).toBeInTheDocument();

    // The old step's dead-end taxonomy must not resurface.
    expect(screen.queryByText('HR Manager')).not.toBeInTheDocument();
    expect(screen.queryByText('Payroll Admin')).not.toBeInTheDocument();
    expect(screen.queryByText('Shift Manager')).not.toBeInTheDocument();
  });

  it('has no editable checkboxes — it is a preview, not a form', () => {
    render(<RolesPermissionsStep onNext={vi.fn()} onBack={vi.fn()} />);
    expect(screen.queryAllByRole('checkbox')).toHaveLength(0);
  });

  it('calls onNext with no arguments when continuing', () => {
    const onNext = vi.fn();
    render(<RolesPermissionsStep onNext={onNext} onBack={vi.fn()} />);

    fireEvent.click(screen.getByRole('button', { name: 'Next' }));

    expect(onNext).toHaveBeenCalledWith();
  });
});
