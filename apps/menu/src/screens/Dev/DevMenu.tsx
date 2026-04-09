import React, { useState } from 'react';
import {
  GridLayout,
  ListLayout,
  MinimalLayout,
  BusinessInfo,
  MenuCategoryData,
  ChatMessage,
} from '@voxire/menu';
import styles from './DevMenu.module.scss';

type LayoutType = 'grid' | 'list' | 'minimal';

const MOCK_BUSINESS: BusinessInfo = {
  name: 'The Golden Fork',
  description: 'Contemporary bistro with a seasonal menu and craft cocktails.',
  address: '12 Maple Street, Downtown',
  phone: '+1 (555) 012-3456',
  hours: 'Mon–Sat 11am–11pm',
};

const MOCK_CATEGORIES: MenuCategoryData[] = [
  {
    id: 'starters',
    name: 'Starters',
    description: 'Small plates to kick things off',
    items: [
      {
        id: 's1',
        name: 'Truffle Fries',
        description: 'Crispy fries tossed in truffle oil, parmesan & fresh herbs.',
        price: 9.99,
        currency: 'USD',
        badges: ['V'],
        available: true,
      },
      {
        id: 's2',
        name: 'Burrata & Heirloom Tomato',
        description: 'Fresh burrata, heirloom tomatoes, basil oil and flaky sea salt.',
        price: 14.5,
        currency: 'USD',
        badges: ['V', 'GF'],
        available: true,
      },
      {
        id: 's3',
        name: 'Crispy Calamari',
        description: 'Lightly battered calamari with lemon aioli and marinara.',
        price: 12.99,
        currency: 'USD',
        available: true,
      },
      {
        id: 's4',
        name: 'Chicken Wings',
        description: 'Eight wings, choice of buffalo, BBQ or garlic parmesan.',
        price: 13.99,
        currency: 'USD',
        badges: ['Spicy'],
        available: true,
      },
    ],
  },
  {
    id: 'mains',
    name: 'Mains',
    description: 'Hearty dishes made from locally sourced ingredients',
    items: [
      {
        id: 'm1',
        name: 'Grilled Salmon',
        description: 'Atlantic salmon, lemon butter, seasonal vegetables and wild rice.',
        price: 26.0,
        currency: 'USD',
        badges: ['GF'],
        available: true,
      },
      {
        id: 'm2',
        name: 'Prime Rib Burger',
        description: '8oz prime rib patty, aged cheddar, caramelised onions, brioche bun.',
        price: 19.5,
        currency: 'USD',
        available: true,
      },
      {
        id: 'm3',
        name: 'Wild Mushroom Risotto',
        description: 'Arborio rice, mixed wild mushrooms, truffle oil, aged parmesan.',
        price: 22.0,
        currency: 'USD',
        badges: ['V'],
        available: true,
      },
      {
        id: 'm4',
        name: 'Chicken Parmigiana',
        description: 'Breaded chicken breast, house tomato sauce, mozzarella, fettuccine.',
        price: 21.0,
        currency: 'USD',
        available: true,
      },
      {
        id: 'm5',
        name: 'NY Strip Steak',
        description: '12oz strip, herb butter, fries, and seasonal greens.',
        price: 42.0,
        currency: 'USD',
        badges: ['GF'],
        available: false,
      },
    ],
  },
  {
    id: 'drinks',
    name: 'Drinks',
    items: [
      {
        id: 'd1',
        name: 'House Lemonade',
        description: 'Fresh squeezed, mint and sparkling water.',
        price: 4.5,
        currency: 'USD',
        badges: ['V', 'GF'],
        available: true,
      },
      {
        id: 'd2',
        name: 'Craft IPA',
        description: 'Local brewery rotating tap. Ask your server for today's selection.',
        price: 7.0,
        currency: 'USD',
        available: true,
      },
      {
        id: 'd3',
        name: 'Cold Brew Coffee',
        description: '24-hour cold brew over ice, oat milk available.',
        price: 5.5,
        currency: 'USD',
        badges: ['V'],
        available: true,
      },
    ],
  },
  {
    id: 'desserts',
    name: 'Desserts',
    items: [
      {
        id: 'des1',
        name: 'Warm Chocolate Fondant',
        description: 'Served with vanilla bean ice cream and berry coulis.',
        price: 9.0,
        currency: 'USD',
        badges: ['V'],
        available: true,
      },
      {
        id: 'des2',
        name: 'Crème Brûlée',
        description: 'Classic vanilla custard with caramelised sugar crust.',
        price: 8.5,
        currency: 'USD',
        badges: ['V', 'GF'],
        available: true,
      },
    ],
  },
];

const ACCENT_OPTIONS = [
  { label: 'Black', value: '#1a1a1a', text: '#ffffff' },
  { label: 'Indigo', value: '#4f46e5', text: '#ffffff' },
  { label: 'Amber', value: '#d97706', text: '#ffffff' },
  { label: 'Rose', value: '#e11d48', text: '#ffffff' },
  { label: 'Teal', value: '#0d9488', text: '#ffffff' },
];

const DevMenu: React.FC = () => {
  const [layout, setLayout] = useState<LayoutType>('grid');
  const [accent, setAccent] = useState(ACCENT_OPTIONS[0]);
  const [enableCart, setEnableCart] = useState(true);
  const [enableAi, setEnableAi] = useState(true);

  // Apply accent to CSS vars live
  React.useEffect(() => {
    document.documentElement.style.setProperty('--menu-accent', accent.value);
    document.documentElement.style.setProperty('--menu-accent-text', accent.text);
  }, [accent]);

  const handleAiMessage = async (message: string, _history: ChatMessage[]): Promise<string> => {
    await new Promise((r) => setTimeout(r, 700));
    return `[Dev mock] You asked: "${message}". Wire onSendMessage to a real AI endpoint to get live responses.`;
  };

  const layoutProps = {
    business: MOCK_BUSINESS,
    categories: MOCK_CATEGORIES,
    enableCart,
    enableAiChat: enableAi,
    onAiMessage: handleAiMessage,
  };

  return (
    <div className={styles.wrapper}>
      {/* Dev toolbar */}
      <div className={styles.toolbar}>
        <span className={styles.toolbar__badge}>DEV</span>

        <div className={styles.toolbar__group}>
          <span className={styles.toolbar__label}>Layout</span>
          {(['grid', 'list', 'minimal'] as LayoutType[]).map((l) => (
            <button
              key={l}
              className={`${styles.toolbar__btn} ${layout === l ? styles['toolbar__btn--active'] : ''}`}
              onClick={() => setLayout(l)}
            >
              {l}
            </button>
          ))}
        </div>

        <div className={styles.toolbar__group}>
          <span className={styles.toolbar__label}>Accent</span>
          {ACCENT_OPTIONS.map((a) => (
            <button
              key={a.value}
              className={`${styles.toolbar__colorBtn} ${accent.value === a.value ? styles['toolbar__colorBtn--active'] : ''}`}
              style={{ background: a.value }}
              onClick={() => setAccent(a)}
              title={a.label}
            />
          ))}
        </div>

        <div className={styles.toolbar__group}>
          <label className={styles.toolbar__toggle}>
            <input type="checkbox" checked={enableCart} onChange={(e) => setEnableCart(e.target.checked)} />
            Cart
          </label>
          <label className={styles.toolbar__toggle}>
            <input type="checkbox" checked={enableAi} onChange={(e) => setEnableAi(e.target.checked)} />
            AI Chat
          </label>
        </div>
      </div>

      {/* Menu */}
      <div className={styles.preview}>
        {layout === 'list' && <ListLayout {...layoutProps} />}
        {layout === 'minimal' && <MinimalLayout {...layoutProps} enableCart={enableCart} />}
        {layout === 'grid' && <GridLayout {...layoutProps} />}
      </div>
    </div>
  );
};

export default DevMenu;
