import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';
import enCommon from '../../../locales/en/common.json';
import enSetup from '../../../locales/en/setup.json';
import RolesPermissionsStep from './RolesPermissionsStep';

// This app's i18n bootstrap (`shared/services/i18next.ts`) loads every
// namespace file into a single 'translation' namespace, keyed by filename, so
// `t('setup.rolesPermissions.title')` resolves as a dot-path within that one
// namespace. Reproduce that here with the real, shipped English copy so the
// assertions below check actual production strings, not raw i18n keys.
void i18next.use(initReactI18next).init({
  lng: 'en',
  fallbackLng: 'en',
  resources: { en: { translation: { common: enCommon, setup: enSetup } } },
  interpolation: { escapeValue: false },
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
