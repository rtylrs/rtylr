import React from 'react';
import { render, screen } from '@testing-library/react';
import { HelmetProvider } from 'react-helmet-async';
import { I18nextProvider } from 'react-i18next';
import { describe, expect, it } from 'vitest';
import i18next from '../../shared/services/i18next';
import RootLanding from './RootLanding';

describe('Recruit RootLanding', () => {
  it('renders English copy instead of raw translation keys with the standalone recruit i18n instance', async () => {
    await i18next.changeLanguage('en');

    render(
      <HelmetProvider>
        <I18nextProvider i18n={i18next}>
          <RootLanding />
        </I18nextProvider>
      </HelmetProvider>
    );

    expect(
      await screen.findByRole('heading', {
        name: /Hiring pages,\s*without the careers-site tax\./i,
      })
    ).toBeVisible();
    expect(screen.getByRole('link', { name: 'See pricing' })).toBeVisible();
    expect(screen.getByRole('link', { name: 'Start free' })).toBeVisible();
    expect(screen.queryByText(/recruitPublic\.landing\./)).not.toBeInTheDocument();
  });
});
