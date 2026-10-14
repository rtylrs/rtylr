import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves,
// which is how this suite passed while every string on screen was a dotted
// path (rtylr#1269). The real locale files mean the copy below also proves the
// lookup.

const helmetSpy = vi.hoisted(() => vi.fn());
vi.mock('@voxire/components', () => ({
  FullScreenLoader: () => <div data-testid="fullscreen-loader" />,
  HelmetContainer: ({ title }: { title: string }) => {
    helmetSpy(title);
    return <title>{title}</title>;
  },
}));

const gridLayoutSpy = vi.hoisted(() => vi.fn());
vi.mock('@voxire/menu', () => ({
  GridLayout: (props: any) => {
    gridLayoutSpy(props);
    return (
      <div data-testid="grid-layout">
        <span data-testid="business-name">{props.business?.name}</span>
        {props.categories?.map((cat: any) => (
          <div key={cat.id} data-testid={`category-${cat.id}`}>
            {cat.items?.map((item: any) => (
              <div key={item.id} data-testid={`item-${item.id}`}>
                <span>{item.name}</span>
                <span>{item.price}</span>
                <span>{item.available ? 'available' : 'unavailable'}</span>
              </div>
            ))}
          </div>
        ))}
        {props.enableCart && (
          <button
            data-testid="mock-checkout"
            onClick={() =>
              props.onCheckout?.({
                items: [
                  {
                    item: { id: '10', name: 'Falafel wrap', price: 5, currency: 'USD' },
                    quantity: 2,
                    modifiers: [],
                  },
                ],
                total: 10,
              })
            }
          >
            Checkout
          </button>
        )}
      </div>
    );
  },
}));

const navigateSpy = vi.hoisted(() => vi.fn());
const backSpy = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ slug: 'acme', location: 'downtown' }),
  useNavigate: () => navigateSpy,
  useRouter: () => ({ history: { back: backSpy } }),
}));

const businessQuery = vi.hoisted(() => ({
  data: undefined as unknown,
  isLoading: false,
  isError: true,
}));
const menuQuery = vi.hoisted(() => ({
  data: undefined as unknown,
  isLoading: false,
  isError: false,
}));
vi.mock('../../../shared/hooks/usePublicMenu', () => ({
  usePublicBusinessBySlug: () => businessQuery,
  usePublicMenu: () => menuQuery,
}));

const cartStoreMock = vi.hoisted(() => ({
  businessId: 0,
  clearCart: vi.fn(),
  setBusinessId: vi.fn((id: number) => {
    cartStoreMock.businessId = id;
  }),
  setLocationId: vi.fn(),
  addItem: vi.fn(),
  updateQuantity: vi.fn(),
}));

vi.mock('../../../shared/stores/cartStore', () => ({
  useCartStore: () => cartStoreMock,
}));

import LocationMenu from './LocationMenu';
import { initTestI18n } from '../../../shared/testing/i18n';

beforeAll(async () => {
  await initTestI18n('en');
});

describe('LocationMenu loading and empty states', () => {
  beforeEach(() => {
    navigateSpy.mockClear();
    backSpy.mockClear();
    gridLayoutSpy.mockClear();
    helmetSpy.mockClear();
    businessQuery.data = undefined;
    businessQuery.isLoading = false;
    businessQuery.isError = true;
    menuQuery.data = undefined;
    menuQuery.isLoading = false;
    menuQuery.isError = false;
  });

  it('renders FullScreenLoader while business or menu is loading', () => {
    businessQuery.isLoading = true;
    businessQuery.isError = false;

    render(<LocationMenu />);
    expect(screen.getByTestId('fullscreen-loader')).toBeInTheDocument();
  });

  it('shows business-not-found and goes back on the CTA', () => {
    render(<LocationMenu />);

    expect(screen.getByText('Business not found')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Go back' }));
    expect(backSpy).toHaveBeenCalledTimes(1);
  });

  it('shows location-not-found and navigates back to the business on the CTA', () => {
    businessQuery.isError = false;
    businessQuery.data = {
      business: { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' },
      locations: [],
    };

    render(<LocationMenu />);

    expect(screen.getByText('Location not found')).toBeInTheDocument();

    const cta = screen.getByRole('button', {
      name: 'Back to Acme Diner',
    });
    fireEvent.click(cta);

    expect(navigateSpy).toHaveBeenCalledWith({ to: '/acme/menu' });
  });

  it('shows a translated message when the menu itself fails to load', () => {
    businessQuery.isError = false;
    businessQuery.data = {
      business: { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' },
      locations: [{ id: 5, name: 'Downtown', slug: 'downtown', address: '', timezone: 'UTC' }],
    };
    menuQuery.isError = true;

    render(<LocationMenu />);

    expect(screen.getByRole('heading', { name: "Couldn't load the menu" })).toBeInTheDocument();
  });

  it('shows no-items-available and navigates back to the business on the CTA', () => {
    businessQuery.isError = false;
    businessQuery.data = {
      business: { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' },
      locations: [{ id: 5, name: 'Downtown', slug: 'downtown', address: '', timezone: 'UTC' }],
    };
    menuQuery.data = {
      business: { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' },
      items: [],
    };

    render(<LocationMenu />);

    expect(screen.getByText('No menu items available')).toBeInTheDocument();

    const cta = screen.getByRole('button', {
      name: 'Back to Acme Diner',
    });
    fireEvent.click(cta);

    expect(navigateSpy).toHaveBeenCalledWith({ to: '/acme/menu' });
  });
});

describe('LocationMenu normal rendering', () => {
  const acmeBusiness = {
    business: {
      id: 1,
      name: 'Acme Diner',
      logoUrl: 'https://example.com/logo.png',
      currency: 'USD',
      timezone: 'UTC',
    },
    locations: [
      { id: 5, name: 'Downtown', slug: 'downtown', address: '123 Main St', timezone: 'UTC' },
    ],
  };
  const acmeMenu = {
    business: {
      id: 1,
      name: 'Acme Diner',
      logoUrl: 'https://example.com/logo.png',
      currency: 'USD',
      timezone: 'UTC',
    },
    items: [
      {
        id: 10,
        name: 'Falafel wrap',
        salePrice: 8.5,
        isActive: true,
        imageUrl: 'https://example.com/wrap.jpg',
      },
      { id: 11, name: 'Seasonal Soup', salePrice: 6.0, isActive: false, imageUrl: undefined },
    ],
  };

  beforeEach(() => {
    businessQuery.isLoading = false;
    businessQuery.isError = false;
    businessQuery.data = acmeBusiness;
    menuQuery.isLoading = false;
    menuQuery.isError = false;
    menuQuery.data = acmeMenu;
    gridLayoutSpy.mockClear();
    helmetSpy.mockClear();
    navigateSpy.mockClear();
  });

  it('renders GridLayout with mapped categories and business info', () => {
    render(<LocationMenu />);

    expect(screen.getByTestId('grid-layout')).toBeInTheDocument();
    expect(screen.getByTestId('business-name')).toHaveTextContent('Acme Diner - Downtown');
    expect(screen.getByText('Falafel wrap')).toBeInTheDocument();
    expect(screen.getByText('available')).toBeInTheDocument();
    expect(screen.getByText('Seasonal Soup')).toBeInTheDocument();
    expect(screen.getByText('unavailable')).toBeInTheDocument();

    expect(helmetSpy).toHaveBeenCalledWith('Menu · rtylr');

    expect(gridLayoutSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        business: {
          name: 'Acme Diner - Downtown',
          logo: 'https://example.com/logo.png',
        },
        categories: [
          {
            id: 'all',
            name: 'Menu',
            items: [
              {
                id: '10',
                name: 'Falafel wrap',
                price: 8.5,
                currency: 'USD',
                description: undefined,
                image: 'https://example.com/wrap.jpg',
                available: true,
              },
              {
                id: '11',
                name: 'Seasonal Soup',
                price: 6.0,
                currency: 'USD',
                description: undefined,
                image: undefined,
                available: false,
              },
            ],
          },
        ],
      })
    );
  });
});

describe('LocationMenu cart reconciliation', () => {
  const acmeBusiness = {
    business: { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' },
    locations: [{ id: 5, name: 'Downtown', slug: 'downtown', address: '', timezone: 'UTC' }],
  };
  const acmeMenu = {
    business: { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' },
    items: [{ id: 10, name: 'Falafel wrap', salePrice: 5, isActive: true }],
  };

  beforeEach(() => {
    businessQuery.isLoading = false;
    businessQuery.isError = false;
    businessQuery.data = acmeBusiness;
    menuQuery.isLoading = false;
    menuQuery.isError = false;
    menuQuery.data = acmeMenu;
    cartStoreMock.businessId = 0;
    cartStoreMock.clearCart.mockClear();
    cartStoreMock.setBusinessId.mockClear();
    cartStoreMock.setLocationId.mockClear();
    cartStoreMock.addItem.mockClear();
  });

  it('clears a persisted cart from another business before adopting this business', () => {
    cartStoreMock.businessId = 99; // abandoned checkout at a different restaurant

    render(<LocationMenu />);

    expect(cartStoreMock.clearCart).toHaveBeenCalledTimes(1);
    expect(cartStoreMock.clearCart.mock.invocationCallOrder[0]).toBeLessThan(
      cartStoreMock.setBusinessId.mock.invocationCallOrder[0]
    );
    expect(cartStoreMock.setBusinessId).toHaveBeenCalledWith(1);
  });

  it('keeps the persisted cart when it already belongs to this business', () => {
    cartStoreMock.businessId = 1;

    render(<LocationMenu />);

    expect(cartStoreMock.clearCart).not.toHaveBeenCalled();
    expect(cartStoreMock.setBusinessId).toHaveBeenCalledWith(1);
  });

  it('does not clear an empty cart', () => {
    cartStoreMock.businessId = 0;

    render(<LocationMenu />);

    expect(cartStoreMock.clearCart).not.toHaveBeenCalled();
  });
});
