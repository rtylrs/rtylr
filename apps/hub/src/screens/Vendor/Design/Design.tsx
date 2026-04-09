import React, { useState } from 'react';
import { X, Plus } from 'phosphor-react';
import {
  MenuLayout,
  MenuLayoutConfig,
  BusinessInfo,
  MenuCategoryData,
  CoverHeight,
  NavStyle,
  NavSize,
  ItemLayout,
  ImagePosition,
  ImageSize,
  CardRadius,
} from '@voxire/menu';
import { useLayoutPresetsStore } from '../../../stores/layoutPresetsStore';

// ─── Mock preview data ─────────────────────────────────────────────────────────

const PH = (w: number, h: number, bg: string, text: string) =>
  `https://placehold.co/${w}x${h}/${bg}/ffffff?text=${encodeURIComponent(text)}`;

const MOCK_BUSINESS: BusinessInfo = {
  name: 'The Golden Fork',
  description: 'Contemporary bistro with a seasonal menu and craft cocktails.',
  coverImage: PH(800, 300, '1a1a1a', 'The+Golden+Fork'),
  logo: PH(120, 120, 'c8a96e', 'GF'),
  address: '12 Maple Street, Downtown',
  phone: '+1 (555) 012-3456',
  hours: 'Mon–Sat 11am–11pm',
  averageRating: 4.6,
  reviewCount: 124,
};

const MOCK_CATEGORIES: MenuCategoryData[] = [
  {
    id: 'starters',
    name: 'Starters',
    items: [
      {
        id: 's1',
        name: 'Truffle Fries',
        description: 'Crispy fries tossed in truffle oil and parmesan.',
        price: 9.99,
        currency: 'USD',
        badges: ['V'],
        available: true,
        image: PH(400, 300, '3d2b1f', 'Truffle+Fries'),
      },
      {
        id: 's2',
        name: 'Burrata & Tomato',
        description: 'Fresh burrata, heirloom tomatoes, basil oil.',
        price: 14.5,
        currency: 'USD',
        badges: ['V', 'GF'],
        available: true,
        image: PH(400, 300, '4a3728', 'Burrata'),
      },
      {
        id: 's3',
        name: 'Crispy Calamari',
        description: 'Lightly battered with lemon aioli.',
        price: 12.99,
        currency: 'USD',
        available: true,
        image: PH(400, 300, '2d3d4a', 'Calamari'),
      },
      {
        id: 's4',
        name: 'Chicken Wings',
        description: 'Eight wings, choice of sauce.',
        price: 13.99,
        currency: 'USD',
        available: true,
        image: PH(400, 300, '4a2d2d', 'Wings'),
      },
    ],
  },
  {
    id: 'mains',
    name: 'Mains',
    items: [
      {
        id: 'm1',
        name: 'Grilled Salmon',
        description: 'Atlantic salmon, lemon butter, seasonal vegetables.',
        price: 26.0,
        currency: 'USD',
        badges: ['GF'],
        available: true,
        image: PH(400, 300, '1e3a2f', 'Salmon'),
      },
      {
        id: 'm2',
        name: 'Classic Burger',
        description: 'Angus beef, cheddar, pickles, house sauce.',
        price: 18.0,
        currency: 'USD',
        available: true,
        image: PH(400, 300, '4a2d1a', 'Burger'),
      },
    ],
  },
];

// ─── Helpers ───────────────────────────────────────────────────────────────────

function BtnGroup<T extends string>({
  value,
  options,
  onChange,
}: {
  value: T;
  options: { label: string; value: T }[];
  onChange: (v: T) => void;
}) {
  return (
    <div className="flex rounded-lg overflow-hidden border border-black/10">
      {options.map((o) => (
        <button
          key={o.value}
          className={`px-3 py-1.5 text-xs font-medium transition-colors ${
            value === o.value
              ? 'bg-admin text-white'
              : 'bg-white text-primary hover:bg-black/5'
          }`}
          onClick={() => onChange(o.value)}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}

function Toggle({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <div className="flex items-center justify-between py-1.5">
      <span className="text-sm text-primary">{label}</span>
      <label className="relative inline-flex items-center cursor-pointer">
        <input
          type="checkbox"
          checked={checked}
          onChange={(e) => onChange(e.target.checked)}
          className="sr-only peer"
        />
        <div className="w-9 h-5 bg-black/15 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-accent" />
      </label>
    </div>
  );
}

// ─── Component ─────────────────────────────────────────────────────────────────

const Design: React.FC = () => {
  const {
    presets,
    activePresetId,
    draftConfig,
    activatePreset,
    createPreset,
    deletePreset,
    patchDraftConfig,
    saveActivePreset,
  } = useLayoutPresetsStore();

  const [newPresetName, setNewPresetName] = useState('');

  const activePreset = presets.find((p) => p.id === activePresetId);

  const patchHeader = (patch: Partial<MenuLayoutConfig['header']>) =>
    patchDraftConfig({ header: { ...draftConfig.header, ...patch } });

  const patchNav = (patch: Partial<MenuLayoutConfig['nav']>) =>
    patchDraftConfig({ nav: { ...draftConfig.nav, ...patch } });

  const patchItems = (patch: Partial<MenuLayoutConfig['items']>) =>
    patchDraftConfig({ items: { ...draftConfig.items, ...patch } });

  const patchPages = (patch: Partial<MenuLayoutConfig['pages']>) =>
    patchDraftConfig({ pages: { ...draftConfig.pages, ...patch } });

  const handleAddPreset = () => {
    const name = newPresetName.trim();
    if (!name) return;
    const id = createPreset(name);
    activatePreset(id);
    setNewPresetName('');
  };

  return (
    <div className="grid h-full overflow-hidden" style={{ gridTemplateColumns: '220px 1fr 380px' }}>
      {/* ── Left: Presets ──────────────────────────────────────────────────── */}
      <aside className="bg-admin flex flex-col h-full overflow-hidden border-r border-white/8">
        <div className="px-4 py-4 border-b border-white/10 shrink-0">
          <h3 className="text-sm font-semibold text-white">Layouts</h3>
        </div>

        <div className="flex-1 overflow-y-auto p-2 space-y-1">
          {presets.map((preset) => (
            <div
              key={preset.id}
              className={`flex items-center gap-2 px-3 py-2 rounded-lg cursor-pointer transition-colors ${
                preset.id === activePresetId
                  ? 'bg-white/15 text-white'
                  : 'text-white/75 hover:bg-white/8'
              }`}
              onClick={() => activatePreset(preset.id)}
            >
              {preset.id === activePresetId && (
                <span className="w-1.5 h-1.5 rounded-full bg-accent shrink-0" />
              )}
              <span className="text-[13px] font-medium flex-1 truncate">{preset.name}</span>
              {presets.length > 1 && (
                <button
                  className="p-0.5 text-white/40 hover:text-white transition-colors"
                  onClick={(e) => {
                    e.stopPropagation();
                    deletePreset(preset.id);
                  }}
                >
                  <X size={12} weight="bold" />
                </button>
              )}
            </div>
          ))}
        </div>

        <div className="flex gap-1.5 p-2 border-t border-white/10 shrink-0">
          <input
            className="flex-1 bg-white/10 text-white placeholder-white/40 text-sm rounded-lg px-3 py-1.5 outline-none focus:bg-white/15 transition-colors"
            placeholder="New layout name…"
            value={newPresetName}
            onChange={(e) => setNewPresetName(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAddPreset()}
          />
          <button
            className="p-2 bg-white/10 text-white rounded-lg hover:bg-white/20 transition-colors disabled:opacity-40"
            onClick={handleAddPreset}
            disabled={!newPresetName.trim()}
          >
            <Plus size={16} weight="bold" />
          </button>
        </div>
      </aside>

      {/* ── Middle: Config controls ─────────────────────────────────────────── */}
      <div className="overflow-y-auto p-5 flex flex-col gap-5 bg-admin-bg">
        {/* Save bar */}
        <div className="flex items-center justify-between bg-white rounded-xl px-5 py-4 border border-black/5 shadow-card">
          <div>
            <h2 className="text-base font-semibold text-primary">{activePreset?.name ?? 'Layout'}</h2>
            <p className="text-xs text-muted mt-0.5">Customize every detail of your menu appearance</p>
          </div>
          <button
            className="px-4 py-2 bg-admin text-white text-sm font-medium rounded-lg hover:opacity-90 transition-opacity"
            onClick={saveActivePreset}
          >
            Save
          </button>
        </div>

        {/* Pages */}
        <div className="bg-white rounded-xl border border-black/5 shadow-card overflow-hidden">
          <div className="px-5 py-3 border-b border-black/5 text-sm font-semibold text-primary">Pages</div>
          <div className="px-5 py-3 space-y-1">
            {(
              [
                ['About', 'about'],
                ['Gallery', 'gallery'],
                ['Reviews', 'reviews'],
                ['Offers', 'offers'],
                ['Book', 'book'],
              ] as [string, keyof MenuLayoutConfig['pages']][]
            ).map(([label, key]) => (
              <Toggle
                key={key}
                label={label}
                checked={draftConfig.pages[key] === true}
                onChange={(v) => patchPages({ [key]: v })}
              />
            ))}
          </div>
        </div>

        {/* Header */}
        <div className="bg-white rounded-xl border border-black/5 shadow-card overflow-hidden">
          <div className="px-5 py-3 border-b border-black/5 text-sm font-semibold text-primary">Header</div>
          <div className="px-5 py-3 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm text-primary">Cover height</span>
              <BtnGroup<CoverHeight>
                value={draftConfig.header.coverHeight}
                options={[
                  { label: 'None', value: 'none' },
                  { label: 'S', value: 'sm' },
                  { label: 'M', value: 'md' },
                  { label: 'L', value: 'lg' },
                ]}
                onChange={(v) => patchHeader({ coverHeight: v })}
              />
            </div>
            <Toggle label="Sticky header" checked={draftConfig.header.sticky} onChange={(v) => patchHeader({ sticky: v })} />
            <Toggle label="Show logo" checked={draftConfig.header.showLogo} onChange={(v) => patchHeader({ showLogo: v })} />
            <Toggle label="Show description" checked={draftConfig.header.showDescription} onChange={(v) => patchHeader({ showDescription: v })} />
            <Toggle label="Show address / phone / hours" checked={draftConfig.header.showMeta} onChange={(v) => patchHeader({ showMeta: v })} />
          </div>
        </div>

        {/* Nav tabs */}
        <div className="bg-white rounded-xl border border-black/5 shadow-card overflow-hidden">
          <div className="px-5 py-3 border-b border-black/5 text-sm font-semibold text-primary">Nav Tabs</div>
          <div className="px-5 py-3 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm text-primary">Style</span>
              <BtnGroup<NavStyle>
                value={draftConfig.nav.style}
                options={[
                  { label: 'Underline', value: 'underline' },
                  { label: 'Pill', value: 'pill' },
                  { label: 'Flat', value: 'flat' },
                ]}
                onChange={(v) => patchNav({ style: v })}
              />
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-primary">Size</span>
              <BtnGroup<NavSize>
                value={draftConfig.nav.size}
                options={[
                  { label: 'S', value: 'sm' },
                  { label: 'M', value: 'md' },
                  { label: 'L', value: 'lg' },
                ]}
                onChange={(v) => patchNav({ size: v })}
              />
            </div>
          </div>
        </div>

        {/* Cards */}
        <div className="bg-white rounded-xl border border-black/5 shadow-card overflow-hidden">
          <div className="px-5 py-3 border-b border-black/5 text-sm font-semibold text-primary">Item Cards</div>
          <div className="px-5 py-3 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm text-primary">Layout</span>
              <BtnGroup<ItemLayout>
                value={draftConfig.items.layout}
                options={[
                  { label: 'Grid', value: 'grid' },
                  { label: 'List', value: 'list' },
                  { label: 'Minimal', value: 'minimal' },
                ]}
                onChange={(v) => patchItems({ layout: v })}
              />
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-primary">Columns</span>
              <BtnGroup<string>
                value={String(draftConfig.items.columns)}
                options={[
                  { label: '2', value: '2' },
                  { label: '3', value: '3' },
                  { label: '4', value: '4' },
                ]}
                onChange={(v) => patchItems({ columns: Number(v) as 2 | 3 | 4 })}
              />
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-primary">Image position</span>
              <BtnGroup<ImagePosition>
                value={draftConfig.items.imagePosition}
                options={[
                  { label: 'Top', value: 'top' },
                  { label: 'Right', value: 'right' },
                  { label: 'Left', value: 'left' },
                  { label: 'None', value: 'none' },
                ]}
                onChange={(v) => patchItems({ imagePosition: v })}
              />
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-primary">Image size</span>
              <BtnGroup<ImageSize>
                value={draftConfig.items.imageSize}
                options={[
                  { label: 'S', value: 'sm' },
                  { label: 'M', value: 'md' },
                  { label: 'L', value: 'lg' },
                ]}
                onChange={(v) => patchItems({ imageSize: v })}
              />
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-primary">Card radius</span>
              <BtnGroup<CardRadius>
                value={draftConfig.items.cardRadius}
                options={[
                  { label: 'None', value: 'none' },
                  { label: 'S', value: 'sm' },
                  { label: 'M', value: 'md' },
                  { label: 'L', value: 'lg' },
                ]}
                onChange={(v) => patchItems({ cardRadius: v })}
              />
            </div>
            <Toggle label="Show description" checked={draftConfig.items.showDescription} onChange={(v) => patchItems({ showDescription: v })} />
            <Toggle label="Show badges" checked={draftConfig.items.showBadges} onChange={(v) => patchItems({ showBadges: v })} />
          </div>
        </div>
      </div>

      {/* ── Right: Live preview ─────────────────────────────────────────────── */}
      <div className="bg-[#e8e8eb] flex flex-col items-center py-6 px-5 overflow-y-auto border-l border-black/8 gap-4">
        <span className="text-xs font-semibold text-black/40 uppercase tracking-widest">Live Preview</span>
        <div className="w-[320px] rounded-[40px] overflow-hidden border-[10px] border-black/80 shadow-2xl bg-white relative">
          <div className="w-24 h-5 bg-black/80 rounded-full absolute top-2 left-1/2 -translate-x-1/2 z-10" />
          <div className="h-[600px] overflow-y-auto overflow-x-hidden">
            <div style={{ transform: 'scale(0.85)', transformOrigin: 'top center', width: '117.6%', marginLeft: '-8.8%' }}>
              <MenuLayout
                business={MOCK_BUSINESS}
                categories={MOCK_CATEGORIES}
                config={draftConfig}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Design;
