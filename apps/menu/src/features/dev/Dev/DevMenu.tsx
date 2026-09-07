import React, { useState, useRef, useEffect } from 'react';
import { cn } from '@voxire/ui';
import { AnimatePresence, motion } from 'motion/react';
import {
  MenuLayout,
  BusinessInfo,
  MenuCategoryData,
  MenuLayoutConfig,
  DEFAULT_LAYOUT_CONFIG,
  ChatMessage,
  CoverHeight,
  NavStyle,
  NavSize,
  ItemLayout,
  ImagePosition,
  ImageSize,
  CardRadius,
} from '@voxire/menu';

// ─── Mock data ─────────────────────────────────────────────────────────────────

const PH = (w: number, h: number, bg: string, text: string) =>
  `https://placehold.co/${w}x${h}/${bg}/ffffff?text=${encodeURIComponent(text)}`;

const MOCK_BUSINESS: BusinessInfo = {
  name: 'The Golden Fork',
  description: 'Contemporary bistro with a seasonal menu and craft cocktails.',
  coverImage: PH(800, 300, '1a1a1a', 'The Golden Fork'),
  logo: PH(120, 120, 'c8a96e', 'GF'),
  address: '12 Maple Street, Downtown',
  phone: '+1 (555) 012-3456',
  email: 'hello@goldenfork.com',
  hours: 'Mon–Sat 11am–11pm',
  website: 'https://goldenfork.com',
  story:
    "Founded in 2010, The Golden Fork started as a small corner bistro with a big dream: to serve honest, seasonal food that brings people together. Today we work with over 30 local farms and producers to bring the best of what's around to your table.",
  socials: {
    instagram: 'thegoldenfork',
    facebook: 'TheGoldenFork',
    tiktok: 'goldenfork',
  },
  photos: [
    PH(600, 600, '2d2d2d', 'Food+1'),
    PH(600, 600, '4a3728', 'Food+2'),
    PH(600, 600, '1e3a2f', 'Food+3'),
    PH(600, 600, '3d2b1f', 'Food+4'),
    PH(600, 600, '1a2a3a', 'Food+5'),
    PH(600, 600, '2a1a3a', 'Food+6'),
    PH(600, 600, '1a3a2a', 'Food+7'),
    PH(600, 600, '3a1a2a', 'Food+8'),
  ],
  averageRating: 4.6,
  reviewCount: 124,
  reviews: [
    {
      id: 'r1',
      author: 'Sarah M.',
      rating: 5,
      comment:
        'Absolutely incredible! The truffle fries alone are worth the trip. Warm atmosphere, attentive staff.',
      date: 'March 2025',
    },
    {
      id: 'r2',
      author: 'James K.',
      rating: 4,
      comment:
        'Great food and a lovely setting. The wild mushroom risotto was perfect. Will definitely be back.',
      date: 'February 2025',
    },
    {
      id: 'r3',
      author: 'Priya N.',
      rating: 5,
      comment:
        'Best burger in the city, hands down. The brioche bun is always fresh and the caramelised onions are chef\u2019s kiss.',
      date: 'January 2025',
    },
  ],
  offers: [
    {
      id: 'o1',
      title: 'Happy Hour: 50% Off Drinks',
      description:
        'Every weekday from 4–6pm. All house cocktails, wine, and draft beers half price.',
      image: PH(800, 350, 'c8a96e', 'Happy+Hour'),
      badge: 'Daily',
      validUntil: '31 Dec 2025',
    },
    {
      id: 'o2',
      title: 'Sunday Brunch Set',
      description:
        '3-course brunch with free-flow mimosas for $45 per person. Reservation required.',
      image: PH(800, 350, '2d4a3a', 'Sunday+Brunch'),
      badge: 'Weekends',
      validUntil: '31 Dec 2025',
    },
    {
      id: 'o3',
      title: 'Loyalty Card: 10th Meal Free',
      description: 'Dine with us 9 times and your 10th main course is completely on the house.',
      badge: 'Ongoing',
    },
  ],
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
        image: PH(400, 300, '3d2b1f', 'Truffle+Fries'),
      },
      {
        id: 's2',
        name: 'Burrata & Heirloom Tomato',
        description: 'Fresh burrata, heirloom tomatoes, basil oil and flaky sea salt.',
        price: 14.5,
        currency: 'USD',
        badges: ['V', 'GF'],
        available: true,
        image: PH(400, 300, '4a3728', 'Burrata'),
      },
      {
        id: 's3',
        name: 'Crispy Calamari',
        description: 'Lightly battered calamari with lemon aioli and marinara.',
        price: 12.99,
        currency: 'USD',
        available: true,
        image: PH(400, 300, '2d3d4a', 'Calamari'),
      },
      {
        id: 's4',
        name: 'Chicken Wings',
        description: 'Eight wings, choice of buffalo, BBQ or garlic parmesan.',
        price: 13.99,
        currency: 'USD',
        badges: ['Spicy'],
        available: true,
        image: PH(400, 300, '4a2d2d', 'Wings'),
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
        image: PH(400, 300, '1e3a2f', 'Salmon'),
      },
      {
        id: 'm2',
        name: 'Prime Rib Burger',
        description: '8oz prime rib patty, aged cheddar, caramelised onions, brioche bun.',
        price: 19.5,
        currency: 'USD',
        available: true,
        image: PH(400, 300, '3d1e1e', 'Burger'),
      },
      {
        id: 'm3',
        name: 'Wild Mushroom Risotto',
        description: 'Arborio rice, mixed wild mushrooms, truffle oil, aged parmesan.',
        price: 22.0,
        currency: 'USD',
        badges: ['V'],
        available: true,
        image: PH(400, 300, '2d2d1e', 'Risotto'),
      },
      {
        id: 'm4',
        name: 'Chicken Parmigiana',
        description: 'Breaded chicken breast, house tomato sauce, mozzarella, fettuccine.',
        price: 21.0,
        currency: 'USD',
        available: true,
        image: PH(400, 300, '4a3728', 'Parmigiana'),
      },
      {
        id: 'm5',
        name: 'NY Strip Steak',
        description: '12oz strip, herb butter, fries, and seasonal greens.',
        price: 42.0,
        currency: 'USD',
        badges: ['GF'],
        available: false,
        image: PH(400, 300, '1a1a1a', 'Steak'),
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
        image: PH(400, 300, '4a4a1e', 'Lemonade'),
      },
      {
        id: 'd2',
        name: 'Craft IPA',
        description: 'Local brewery rotating tap. Ask your server for today\u2019s selection.',
        price: 7.0,
        currency: 'USD',
        available: true,
        image: PH(400, 300, '2d3d1e', 'IPA'),
      },
      {
        id: 'd3',
        name: 'Cold Brew Coffee',
        description: '24-hour cold brew over ice, oat milk available.',
        price: 5.5,
        currency: 'USD',
        badges: ['V'],
        available: true,
        image: PH(400, 300, '1e1e1e', 'Cold+Brew'),
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
        image: PH(400, 300, '3d1e3d', 'Fondant'),
      },
      {
        id: 'des2',
        name: 'Cr\u00e8me Br\u00fcl\u00e9e',
        description: 'Classic vanilla custard with caramelised sugar crust.',
        price: 8.5,
        currency: 'USD',
        badges: ['V', 'GF'],
        available: true,
        image: PH(400, 300, '4a3d1e', 'Creme+Brulee'),
      },
    ],
  },
];

// ─── Accent presets ────────────────────────────────────────────────────────────

const ACCENT_OPTIONS = [
  { label: 'Ink', value: '#1a1a1a', text: '#ffffff' },
  { label: 'Indigo', value: '#4f46e5', text: '#ffffff' },
  { label: 'Amber', value: '#d97706', text: '#ffffff' },
  { label: 'Rose', value: '#e11d48', text: '#ffffff' },
  { label: 'Teal', value: '#0d9488', text: '#ffffff' },
  { label: 'Sage', value: '#4a7c59', text: '#ffffff' },
];

// ─── Helper components ─────────────────────────────────────────────────────────

const BtnGroup = <T extends string | number>({
  options,
  value,
  onChange,
}: {
  options: { label: string; value: T }[];
  value: T;
  onChange: (v: T) => void;
}) => (
  <div className="flex rounded-lg overflow-hidden border border-black/10">
    {options.map((o) => (
      <button
        type="button"
        key={String(o.value)}
        className={cn(
          'px-2.5 py-1 text-xs font-medium transition-colors',
          value === o.value ? 'bg-gray-900 text-white' : 'bg-white text-gray-600 hover:bg-gray-50'
        )}
        onClick={() => onChange(o.value)}
      >
        {o.label}
      </button>
    ))}
  </div>
);

const Toggle = ({
  checked,
  onChange,
  label,
}: {
  checked: boolean;
  onChange: (v: boolean) => void;
  label: string;
}) => (
  <label className="flex items-center justify-between cursor-pointer py-0.5">
    <span className="text-xs text-gray-600">{label}</span>
    <div className="relative">
      <input
        type="checkbox"
        className="sr-only peer"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
      />
      <div className="w-8 h-4 bg-gray-200 rounded-full peer peer-checked:bg-gray-900 after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:w-3 after:h-3 after:rounded-full after:bg-white after:transition-all peer-checked:after:translate-x-4" />
    </div>
  </label>
);

// ─── Main component ────────────────────────────────────────────────────────────

const DevMenu: React.FC = () => {
  const [open, setOpen] = useState(false);
  const [accent, setAccent] = useState(ACCENT_OPTIONS[0]);
  const [enableCart, setEnableCart] = useState(true);
  const [enableAi, setEnableAi] = useState(true);
  const panelRef = useRef<HTMLDivElement>(null);

  // Full layout config state
  const [cfg, setCfg] = useState<MenuLayoutConfig>({
    ...DEFAULT_LAYOUT_CONFIG,
    pages: { menu: true, about: true, gallery: false, reviews: false, offers: false },
    header: {
      coverHeight: 'md',
      showLogo: true,
      showDescription: true,
      showMeta: true,
      sticky: true,
    },
    nav: { style: 'underline', size: 'md' },
    items: {
      layout: 'grid',
      columns: 2,
      imagePosition: 'top',
      imageSize: 'md',
      showDescription: true,
      showBadges: true,
      cardRadius: 'md',
    },
  });

  const patchHeader = (partial: Partial<MenuLayoutConfig['header']>) =>
    setCfg((prev) => ({ ...prev, header: { ...prev.header, ...partial } }));
  const patchNav = (partial: Partial<MenuLayoutConfig['nav']>) =>
    setCfg((prev) => ({ ...prev, nav: { ...prev.nav, ...partial } }));
  const patchItems = (partial: Partial<MenuLayoutConfig['items']>) =>
    setCfg((prev) => ({ ...prev, items: { ...prev.items, ...partial } }));
  const patchPages = (partial: Partial<MenuLayoutConfig['pages']>) =>
    setCfg((prev) => ({ ...prev, pages: { ...prev.pages, ...partial } }));

  useEffect(() => {
    document.documentElement.style.setProperty('--menu-accent', accent.value);
    document.documentElement.style.setProperty('--menu-accent-text', accent.text);
  }, [accent]);

  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [open]);

  const handleAiMessage = async (message: string, _history: ChatMessage[]): Promise<string> => {
    await new Promise((r) => setTimeout(r, 700));
    return `[Dev mock] You asked: "${message}". Wire onAiMessage to a real endpoint for live responses.`;
  };

  const Section = ({ label, children }: { label: string; children: React.ReactNode }) => (
    <div className="space-y-2">
      <span className="text-[10px] font-bold uppercase tracking-wider text-gray-400">{label}</span>
      {children}
    </div>
  );

  const SubLabel = ({ children }: { children: React.ReactNode }) => (
    <span className="text-[10px] text-gray-400 mt-1">{children}</span>
  );

  return (
    <div className="relative">
      <MenuLayout
        business={MOCK_BUSINESS}
        categories={MOCK_CATEGORIES}
        config={cfg}
        enableCart={enableCart}
        enableAiChat={enableAi}
        onAiMessage={handleAiMessage}
      />

      {/* Floating toolbar */}
      <div ref={panelRef} className="fixed bottom-6 left-6 z-[999]">
        <AnimatePresence>
          {open && (
            <motion.div
              className="absolute bottom-14 left-0 w-64 bg-white border border-black/10 rounded-2xl shadow-xl p-4 space-y-4 max-h-[80vh] overflow-y-auto"
              initial={{ opacity: 0, scale: 0.92, y: 8 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.92, y: 8 }}
              transition={{ type: 'spring', damping: 24, stiffness: 320 }}
            >
              <Section label="Pages">
                <div className="space-y-1">
                  <Toggle
                    checked={!!cfg.pages.about}
                    onChange={(v) => patchPages({ about: v })}
                    label="About"
                  />
                  <Toggle
                    checked={!!cfg.pages.gallery}
                    onChange={(v) => patchPages({ gallery: v })}
                    label="Gallery"
                  />
                  <Toggle
                    checked={!!cfg.pages.reviews}
                    onChange={(v) => patchPages({ reviews: v })}
                    label="Reviews"
                  />
                  <Toggle
                    checked={!!cfg.pages.offers}
                    onChange={(v) => patchPages({ offers: v })}
                    label="Offers"
                  />
                </div>
              </Section>

              <div className="h-px bg-black/8" />

              <Section label="Header">
                <SubLabel>Cover height</SubLabel>
                <BtnGroup<CoverHeight>
                  options={[
                    { label: 'None', value: 'none' },
                    { label: 'S', value: 'sm' },
                    { label: 'M', value: 'md' },
                    { label: 'L', value: 'lg' },
                  ]}
                  value={cfg.header.coverHeight}
                  onChange={(v) => patchHeader({ coverHeight: v })}
                />
                <div className="space-y-1 mt-1">
                  <Toggle
                    checked={cfg.header.sticky}
                    onChange={(v) => patchHeader({ sticky: v })}
                    label="Sticky"
                  />
                  <Toggle
                    checked={cfg.header.showLogo}
                    onChange={(v) => patchHeader({ showLogo: v })}
                    label="Logo"
                  />
                  <Toggle
                    checked={cfg.header.showDescription}
                    onChange={(v) => patchHeader({ showDescription: v })}
                    label="Description"
                  />
                  <Toggle
                    checked={cfg.header.showMeta}
                    onChange={(v) => patchHeader({ showMeta: v })}
                    label="Address / Hours"
                  />
                </div>
              </Section>

              <div className="h-px bg-black/8" />

              <Section label="Nav Tabs">
                <SubLabel>Style</SubLabel>
                <BtnGroup<NavStyle>
                  options={[
                    { label: 'Underline', value: 'underline' },
                    { label: 'Pill', value: 'pill' },
                    { label: 'Flat', value: 'flat' },
                  ]}
                  value={cfg.nav.style}
                  onChange={(v) => patchNav({ style: v })}
                />
                <SubLabel>Size</SubLabel>
                <BtnGroup<NavSize>
                  options={[
                    { label: 'S', value: 'sm' },
                    { label: 'M', value: 'md' },
                    { label: 'L', value: 'lg' },
                  ]}
                  value={cfg.nav.size}
                  onChange={(v) => patchNav({ size: v })}
                />
              </Section>

              <div className="h-px bg-black/8" />

              <Section label="Cards">
                <SubLabel>Layout</SubLabel>
                <BtnGroup<ItemLayout>
                  options={[
                    { label: 'Grid', value: 'grid' },
                    { label: 'List', value: 'list' },
                    { label: 'Minimal', value: 'minimal' },
                  ]}
                  value={cfg.items.layout}
                  onChange={(v) => patchItems({ layout: v })}
                />
                {cfg.items.layout === 'grid' && (
                  <>
                    <SubLabel>Columns</SubLabel>
                    <BtnGroup<number>
                      options={[
                        { label: '2', value: 2 },
                        { label: '3', value: 3 },
                        { label: '4', value: 4 },
                      ]}
                      value={cfg.items.columns}
                      onChange={(v) => patchItems({ columns: v as 2 | 3 | 4 })}
                    />
                  </>
                )}
                {cfg.items.layout !== 'minimal' && (
                  <>
                    <SubLabel>Image position</SubLabel>
                    <BtnGroup<ImagePosition>
                      options={[
                        { label: 'Top', value: 'top' },
                        { label: 'Right', value: 'right' },
                        { label: 'Left', value: 'left' },
                        { label: 'None', value: 'none' },
                      ]}
                      value={cfg.items.imagePosition}
                      onChange={(v) => patchItems({ imagePosition: v })}
                    />
                    <SubLabel>Image size</SubLabel>
                    <BtnGroup<ImageSize>
                      options={[
                        { label: 'S', value: 'sm' },
                        { label: 'M', value: 'md' },
                        { label: 'L', value: 'lg' },
                      ]}
                      value={cfg.items.imageSize}
                      onChange={(v) => patchItems({ imageSize: v })}
                    />
                  </>
                )}
                <SubLabel>Corner radius</SubLabel>
                <BtnGroup<CardRadius>
                  options={[
                    { label: '0', value: 'none' },
                    { label: 'S', value: 'sm' },
                    { label: 'M', value: 'md' },
                    { label: 'L', value: 'lg' },
                  ]}
                  value={cfg.items.cardRadius}
                  onChange={(v) => patchItems({ cardRadius: v })}
                />
                <div className="space-y-1 mt-1">
                  <Toggle
                    checked={cfg.items.showDescription}
                    onChange={(v) => patchItems({ showDescription: v })}
                    label="Description"
                  />
                  <Toggle
                    checked={cfg.items.showBadges}
                    onChange={(v) => patchItems({ showBadges: v })}
                    label="Badges"
                  />
                </div>
              </Section>

              <div className="h-px bg-black/8" />

              <Section label="Accent">
                <div className="flex gap-2 flex-wrap">
                  {ACCENT_OPTIONS.map((a) => (
                    <button
                      type="button"
                      key={a.value}
                      className={cn(
                        'w-7 h-7 rounded-full transition-all',
                        accent.value === a.value
                          ? 'ring-2 ring-offset-1 ring-gray-900 scale-110'
                          : 'hover:scale-105'
                      )}
                      style={{ background: a.value }}
                      onClick={() => setAccent(a)}
                      title={a.label}
                    />
                  ))}
                </div>
              </Section>

              <div className="h-px bg-black/8" />

              <Section label="Features">
                <div className="space-y-1">
                  <Toggle checked={enableCart} onChange={setEnableCart} label="Cart" />
                  <Toggle checked={enableAi} onChange={setEnableAi} label="AI Chat" />
                </div>
              </Section>
            </motion.div>
          )}
        </AnimatePresence>

        <button
          type="button"
          className="w-12 h-12 rounded-full bg-gray-900 text-white shadow-lg hover:opacity-90 transition-opacity flex items-center justify-center text-lg relative"
          onClick={() => setOpen((v) => !v)}
          aria-label="Toggle dev toolbar"
        >
          {open ? '✕' : '⚙'}
          <span className="absolute -top-1 -right-1 text-[8px] font-bold bg-amber-400 text-gray-900 rounded-full px-1 py-0.5 leading-none">
            DEV
          </span>
        </button>
      </div>
    </div>
  );
};

export default DevMenu;
