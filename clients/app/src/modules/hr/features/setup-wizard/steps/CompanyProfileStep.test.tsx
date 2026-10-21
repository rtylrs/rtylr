import { createInstance } from 'i18next';
import { I18nextProvider } from 'react-i18next';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import CompanyProfileStep from './CompanyProfileStep';
import enCommon from '../../../locales/en/hrCommon.json';
import enSetup from '../../../locales/en/hrSetup.json';
import frCommon from '../../../locales/fr/hrCommon.json';
import frSetup from '../../../locales/fr/hrSetup.json';
import arCommon from '../../../locales/ar/hrCommon.json';
import arSetup from '../../../locales/ar/hrSetup.json';

const resources = {
  en: { hrCommon: { hrCommon: enCommon, hrSetup: enSetup } },
  fr: { hrCommon: { hrCommon: frCommon, hrSetup: frSetup } },
  ar: { hrCommon: { hrCommon: arCommon, hrSetup: arSetup } },
};

describe('HR company identity', () => {
  it.each(['en', 'fr', 'ar'] as const)(
    'inherits the shared name and preserves HR options in %s',
    async (lng) => {
      const i18n = createInstance();
      await i18n.init({
        lng,
        resources,
        defaultNS: 'hrCommon',
        interpolation: { escapeValue: false },
      });
      const onNext = vi.fn();
      const { rerender } = render(
        <I18nextProvider i18n={i18n}>
          <CompanyProfileStep
            isFirst
            onNext={onNext}
            sharedName="Northwind"
            initialValues={{ name: 'Old HR name', timezone: 'Europe/Berlin', currency: 'EUR' }}
          />
        </I18nextProvider>
      );
      const copy = resources[lng].hrCommon;
      const name = screen.getByLabelText(copy.hrSetup.companyProfile.name);
      expect(name).toHaveValue('Northwind');
      expect(name).toHaveAttribute('readonly');
      expect(screen.getByText(copy.hrSetup.companyProfile.sharedName)).toBeVisible();
      rerender(
        <I18nextProvider i18n={i18n}>
          <CompanyProfileStep
            isFirst
            onNext={onNext}
            sharedName="Renamed business"
            initialValues={{ name: 'Old HR name' }}
          />
        </I18nextProvider>
      );
      expect(name).toHaveValue('Renamed business');
      fireEvent.click(screen.getByRole('button', { name: copy.hrCommon.next }));
      expect(onNext).toHaveBeenCalledWith({
        name: 'Renamed business',
        timezone: 'Europe/Berlin',
        currency: 'EUR',
      });
    }
  );
  it('keeps legacy setup editable when there is no shared identity', async () => {
    const i18n = createInstance();
    await i18n.init({ lng: 'en', resources, defaultNS: 'hrCommon' });
    const onNext = vi.fn();
    render(
      <I18nextProvider i18n={i18n}>
        <CompanyProfileStep
          isFirst
          onNext={onNext}
          initialValues={{ name: 'Legacy company', timezone: 'Europe/Berlin', currency: 'EUR' }}
        />
      </I18nextProvider>
    );
    const input = screen.getByLabelText('Company name');
    expect(input).not.toHaveAttribute('readonly');
    fireEvent.change(input, { target: { value: 'Updated legacy company' } });
    fireEvent.click(screen.getByRole('button', { name: 'Next' }));
    expect(onNext).toHaveBeenCalledWith({
      name: 'Updated legacy company',
      timezone: 'Europe/Berlin',
      currency: 'EUR',
    });
  });
});
