import React, { useEffect, useState } from 'react';
import { useParams } from '@tanstack/react-router';
import { FullScreenLoader, HelmetContainer } from '@voxire/components';
import { cdnImageUrl } from '@voxire/config';
import {
  GridLayout,
  ListLayout,
  MinimalLayout,
  MenuCategoryData,
  BusinessInfo,
  ChatMessage,
} from '@voxire/menu';
import useAsync from '../../../shared/hooks/useAsync';
import { menuApi } from '../../../shared/api/api';

type LayoutType = 'grid' | 'list' | 'minimal';

interface ThemeConfig {
  layout?: LayoutType;
  primaryColor?: string;
  backgroundColor?: string;
  menuAccent?: string;
}

interface RawMenuItem {
  _id?: string;
  name?: string;
  description?: string;
  price?: number;
  image?: string;
  badges?: string[];
  available?: boolean;
}
interface RawCategory {
  _id?: string;
  type?: string;
  name?: string;
  description?: string;
  menuItems?: RawMenuItem[];
}

function transformCategories(rawItems: RawCategory[]): MenuCategoryData[] {
  return rawItems
    .filter((item) => item.type === 'category' && Array.isArray(item.menuItems))
    .map((cat) => ({
      id: cat._id ?? '',
      name: cat.name ?? '',
      description: cat.description,
      items: (cat.menuItems ?? []).map((item) => ({
        id: item._id ?? '',
        name: item.name ?? '',
        description: item.description,
        price: item.price ?? 0,
        image: cdnImageUrl(item.image),
        currency: 'USD',
        badges: item.badges ?? [],
        available: item.available !== false,
      })),
    }))
    .filter((cat) => cat.items.length > 0);
}

function applyTheme(theme: ThemeConfig) {
  const root = document.documentElement;
  if (theme.primaryColor) {
    root.style.setProperty('--menu-accent', theme.primaryColor);
  }
  if (theme.backgroundColor) {
    root.style.setProperty('--menu-bg', theme.backgroundColor);
  }
}

interface MenuProps {
  id?: string;
}

const Menu: React.FC<MenuProps> = ({ id: propId }) => {
  const params = useParams({ strict: false });
  const id = propId || params.id || params.slug;
  const [categories, setCategories] = useState<MenuCategoryData[]>([]);
  const [business, setBusiness] = useState<BusinessInfo | null>(null);
  const [layout, setLayout] = useState<LayoutType>('grid');
  const [status, setStatus] = useState<'idle' | 'loading' | 'succeeded' | 'failed'>('idle');
  const [retryCount, setRetryCount] = useState(0);

  const getMenu = useAsync({
    fn: menuApi.getMenu,
    onSuccess: (response) => {
      setRetryCount(0);
      const { details } = response;
      const theme: ThemeConfig = details.theme ?? {};

      const cats = transformCategories(details.menu?.items ?? []);
      setCategories(cats);

      setBusiness({
        name: details.business?.name ?? 'Menu',
        logo: cdnImageUrl(details.business?.logo),
        coverImage: cdnImageUrl(details.business?.coverImage),
        description: details.business?.description,
        address: details.business?.address,
        phone: details.business?.phone,
        hours: details.business?.hours,
      });

      setLayout(theme.layout ?? 'grid');
      applyTheme(theme);

      setStatus('succeeded');
    },
    onError: () => {
      setStatus('failed');
      if (retryCount < 3) {
        setTimeout(() => {
          getMenu.main({ id });
          setRetryCount((r) => r + 1);
        }, 8000);
      }
    },
    callOnMount: false,
  });
  const { main: loadMenu } = getMenu;

  useEffect(() => {
    setStatus('loading');
    loadMenu({ id });
  }, [id, loadMenu]);

  // AI chat handler, wired to backend when ready
  const handleAiMessage = async (message: string, _history: ChatMessage[]): Promise<string> => {
    // Placeholder until AI backend endpoint is created
    return `Thanks for your question about "${message}". Our team will have AI-powered menu assistance available soon!`;
  };

  if (status === 'loading' || status === 'idle') {
    return <FullScreenLoader />;
  }

  if (status === 'failed') {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center gap-4 bg-white">
        <h2 className="text-lg font-semibold text-gray-800">Failed to load menu</h2>
        <p className="text-sm text-gray-500">Please try refreshing the page.</p>
        <button
          className="px-4 py-2 text-sm bg-gray-900 text-white rounded-lg hover:opacity-90 transition-opacity"
          onClick={() => {
            setStatus('loading');
            loadMenu({ id });
          }}
        >
          Retry
        </button>
      </div>
    );
  }

  if (!business) return null;

  const handleCheckout = () => {
    navigate({ to: '/checkout' });
  };

  const handleReviewSubmit = async (review: { rating: number; name: string; comment: string }) => {
    try {
      await api.post('/crm/feedback/v1/feedback/public', {
        rating: review.rating,
        name: review.name,
        comment: review.comment,
      });
    } catch {
      // ignore
    }
  };

  const layoutProps = {
    business,
    categories,
    enableCart: true,
    enableAiChat: true,
    onAiMessage: handleAiMessage,
    onCheckout: handleCheckout,
    onSubmitReview: handleReviewSubmit,
  };

  const content =
    layout === 'list' ? (
      <ListLayout {...layoutProps} />
    ) : layout === 'minimal' ? (
      <MinimalLayout {...layoutProps} enableCart={false} />
    ) : (
      <GridLayout {...layoutProps} />
    );

  return (
    <>
      <HelmetContainer title={business.name} />
      {content}
    </>
  );
};

export default Menu;
