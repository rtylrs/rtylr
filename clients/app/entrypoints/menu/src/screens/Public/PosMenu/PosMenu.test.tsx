import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

// i18next is deliberately not mocked, and these assertions name the copy rather
// than the key, for the reason Checkout.test.tsx states. The defect this file
// exists for was invisible to every other test: the eight `posmenu.*` keys had
// no locale file to live in, i18next fell back to echoing the key, and a diner
// on `/b/:id` read "posmenu.loadFailedTitle" where a sentence should be. A test
// that asserted on keys would have agreed with the component and passed.

vi.mock('@voxire/components', () => ({
  FullScreenLoader: () => <div>loading</div>,
  HelmetContainer: ({ title }: { title: string }) => <title>{title}</title>,
}));

vi.mock('@voxire/menu', () => ({
  GridLayout: ({
    categories,
    business,
  }: {
    categories: { name: string }[];
    business: { name: string } | null;
  }) => (
    <div>
      <p>{business?.name}</p>
      {categories.map((c) => (
        <p key={c.id}>{c.name}</p>
      ))}
    </div>
  ),
}));

const routerParams = vi.hoisted(() => ({ businessId: '42' as string | undefined }));
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ businessId: routerParams.businessId }),
  useNavigate: () => vi.fn(),
}));

const menuQuery = vi.hoisted(() => ({
  data: undefined as unknown,
  isLoading: false,
  isError: false,
}));
vi.mock('../../../shared/hooks/usePublicMenu', () => ({
  usePublicMenu: () => menuQuery,
}));

import PosMenu from './PosMenu';
import { useCartStore } from '../../../shared/stores/cartStore';
import { initTestI18n } from '../../../shared/testing/i18n';

/** The copy a diner must see, per language, for each state the screen can land in. */
const COPY = {
  en: {
    invalidUrlTitle: "This menu link isn't valid",
    invalidUrlDescription: 'Check the link in your QR code and try again.',
    loadFailedTitle: "Couldn't load the menu",
    loadFailedDescription: 'Please refresh the page or try again in a few minutes.',
    noItemsTitle: 'No menu items available',
    noItemsDescription: 'This menu is currently empty. Please check back later.',
    allItems: 'Menu',
  },
  fr: {
    invalidUrlTitle: "Ce lien de menu n'est pas valide",
    invalidUrlDescription: 'Vérifiez le lien de votre QR code et réessayez.',
    loadFailedTitle: 'Impossible de charger le menu',
    loadFailedDescription: 'Veuillez actualiser la page ou réessayer dans quelques instants.',
    noItemsTitle: 'Aucun article au menu',
    noItemsDescription: 'Ce menu est actuellement vide. Veuillez revenir plus tard.',
    allItems: 'Menu',
  },
  ar: {
    invalidUrlTitle: 'رابط القائمة غير صالح',
    invalidUrlDescription:
      'يرجى التحقق من الرابط في رمز الاستجابة السريعة (QR) والمحاولة مرة أخرى.',
    loadFailedTitle: 'تعذّر تحميل القائمة',
    loadFailedDescription: 'يرجى تحديث الصفحة أو المحاولة مرة أخرى بعد قليل.',
    noItemsTitle: 'لا توجد عناصر في القائمة',
    noItemsDescription: 'هذه القائمة فارغة حاليًا. يرجى المحاولة مرة أخرى لاحقًا.',
    allItems: 'القائمة',
  },
} as const;

const business = {
  id: 42,
  name: 'Acme Kitchen',
  logoUrl: '',
  currency: 'USD',
  timezone: 'UTC',
};
const item = {
  id: 7,
  ownerId: 1,
  categoryId: 1,
  name: 'Falafel wrap',
  sku: 'FW',
  salePrice: 6.5,
  isActive: true,
  createdAt: '',
  updatedAt: '',
};

/**
 * The screen must never paint a key. Checked on every rendered state in every
 * language, because the symptom is language-independent: the file was missing
 * from all three, so all three shipped dotted paths.
 */
function expectNoRawKeys(container: HTMLElement): void {
  const text = container.textContent ?? '';
  const leaked = text.match(/posmenu\.[A-Za-z0-9_.]+/g);
  expect(leaked ?? []).toEqual([]);
}

const LANGS = ['en', 'fr', 'ar'] as const;

describe.each(LANGS)('PosMenu renders customer copy in %s', (lang) => {
  beforeEach(async () => {
    // The loader is captured per init, so the language is chosen by
    // re-initialising rather than by changeLanguage.
    await initTestI18n(lang);
    useCartStore.getState().clearCart();
    routerParams.businessId = '42';
    menuQuery.data = undefined;
    menuQuery.isLoading = false;
    menuQuery.isError = false;
  });

  it('explains an unusable link without naming the key', () => {
    routerParams.businessId = undefined;

    const { container } = render(<PosMenu />);

    expect(screen.getByText(COPY[lang].invalidUrlTitle)).toBeInTheDocument();
    expect(screen.getByText(COPY[lang].invalidUrlDescription)).toBeInTheDocument();
    expectNoRawKeys(container);
  });

  it('explains a failed load without naming the key', () => {
    menuQuery.isError = true;

    const { container } = render(<PosMenu />);

    expect(screen.getByText(COPY[lang].loadFailedTitle)).toBeInTheDocument();
    expect(screen.getByText(COPY[lang].loadFailedDescription)).toBeInTheDocument();
    expectNoRawKeys(container);
  });

  it('explains an empty menu without naming the key', () => {
    menuQuery.data = { business, items: [] };

    const { container } = render(<PosMenu />);

    expect(screen.getByText(COPY[lang].noItemsTitle)).toBeInTheDocument();
    expect(screen.getByText(COPY[lang].noItemsDescription)).toBeInTheDocument();
    expectNoRawKeys(container);
  });

  it('labels the synthesised category in the language, not as a key', () => {
    menuQuery.data = { business, items: [item] };

    const { container } = render(<PosMenu />);

    expect(screen.getByText(COPY[lang].allItems)).toBeInTheDocument();
    expect(screen.getByText('Acme Kitchen')).toBeInTheDocument();
    expectNoRawKeys(container);
  });
});
