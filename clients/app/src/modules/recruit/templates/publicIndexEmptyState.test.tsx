import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

import EditorialIndex from './editorial/Index';
import MinimalIndex from './minimal/Index';
import BoldIndex from './bold/Index';
import type { IndexProps } from './resolver';
import type { Customization } from '../shared/lib/types';

/**
 * The three public template variants (editorial/minimal/bold) each
 * duplicate the "no open roles" empty state. This test drives all
 * three through the same table so a future edit that updates one
 * template's copy or CTA without the others gets caught here rather
 * than in a live diff review.
 */

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const fakeCustomization: Customization = {
  id: 1,
  slug: 'acme',
  brandName: 'Acme',
  tagline: '',
  aboutBlurb: '',
  logoUrl: '',
  primaryColor: '',
  accentColor: '',
  bgColor: '',
  textColor: '',
  mutedTextColor: '',
  fontPairing: '',
  heroImageUrl: '',
  faviconUrl: '',
  template: 'minimal',
  headerHtml: '',
  metaTitle: '',
  metaDescription: '',
  privacyUrl: '',
  termsUrl: '',
  gaMeasurementId: '',
  metaPixelId: '',
  requireConsent: false,
};

const baseProps: IndexProps = {
  careers: [],
  customization: fakeCustomization,
  sections: [],
};

const TEMPLATES: Array<{ name: string; Component: React.FC<IndexProps> }> = [
  { name: 'editorial', Component: EditorialIndex },
  { name: 'minimal', Component: MinimalIndex },
  { name: 'bold', Component: BoldIndex },
];

describe.each(TEMPLATES)('$name public Index empty state', ({ Component }) => {
  const reloadMock = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { ...window.location, reload: reloadMock },
    });
  });

  it('shows the no-open-roles message with a check-again CTA', () => {
    render(<Component {...baseProps} />);

    expect(screen.getByText('recruitCommon.emptyStates.publicRolesNone')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'recruitCommon.emptyStates.checkAgain' }));
    expect(reloadMock).toHaveBeenCalledTimes(1);
  });
});
