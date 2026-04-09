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
import styles from './Design.module.scss';

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
    <div className={styles.btnGroup}>
      {options.map((o) => (
        <button
          key={o.value}
          className={`${styles.btn} ${value === o.value ? styles.btnActive : ''}`}
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
    <div className={styles.switchRow}>
      <span className={styles.switchLabel}>{label}</span>
      <label className={styles.switch}>
        <input type="checkbox" checked={checked} onChange={(e) => onChange(e.target.checked)} />
        <span className={styles.switchTrack} />
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
    <div className={styles.page}>
      {/* ── Left: Presets ──────────────────────────────────────────────────── */}
      <aside className={styles.presets}>
        <div className={styles.presetsHeader}>
          <h3>Layouts</h3>
        </div>

        <div className={styles.presetList}>
          {presets.map((preset) => (
            <div
              key={preset.id}
              className={`${styles.presetItem} ${preset.id === activePresetId ? styles.active : ''}`}
              onClick={() => activatePreset(preset.id)}
            >
              {preset.id === activePresetId && <span className={styles.presetActiveDot} />}
              <span className={styles.presetName}>{preset.name}</span>
              {presets.length > 1 && (
                <button
                  className={styles.presetDeleteBtn}
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

        <div className={styles.addPresetRow}>
          <input
            className={styles.addPresetInput}
            placeholder="New layout name…"
            value={newPresetName}
            onChange={(e) => setNewPresetName(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAddPreset()}
          />
          <button
            className={styles.addPresetBtn}
            onClick={handleAddPreset}
            disabled={!newPresetName.trim()}
          >
            <Plus size={16} weight="bold" />
          </button>
        </div>
      </aside>

      {/* ── Middle: Config controls ─────────────────────────────────────────── */}
      <div className={styles.controls}>
        {/* Save bar */}
        <div className={styles.saveBar}>
          <div className={styles.saveBarLeft}>
            <h2>{activePreset?.name ?? 'Layout'}</h2>
            <p>Customize every detail of your menu appearance</p>
          </div>
          <button className={styles.saveBtn} onClick={saveActivePreset}>
            Save
          </button>
        </div>

        {/* Pages */}
        <div className={styles.card}>
          <div className={styles.cardHeader}>Pages</div>
          <div className={styles.cardBody}>
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
        <div className={styles.card}>
          <div className={styles.cardHeader}>Header</div>
          <div className={styles.cardBody}>
            <div className={styles.row}>
              <span className={styles.label}>Cover height</span>
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
            <Toggle
              label="Sticky header"
              checked={draftConfig.header.sticky}
              onChange={(v) => patchHeader({ sticky: v })}
            />
            <Toggle
              label="Show logo"
              checked={draftConfig.header.showLogo}
              onChange={(v) => patchHeader({ showLogo: v })}
            />
            <Toggle
              label="Show description"
              checked={draftConfig.header.showDescription}
              onChange={(v) => patchHeader({ showDescription: v })}
            />
            <Toggle
              label="Show address / phone / hours"
              checked={draftConfig.header.showMeta}
              onChange={(v) => patchHeader({ showMeta: v })}
            />
          </div>
        </div>

        {/* Nav tabs */}
        <div className={styles.card}>
          <div className={styles.cardHeader}>Nav Tabs</div>
          <div className={styles.cardBody}>
            <div className={styles.row}>
              <span className={styles.label}>Style</span>
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
            <div className={styles.row}>
              <span className={styles.label}>Size</span>
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
        <div className={styles.card}>
          <div className={styles.cardHeader}>Item Cards</div>
          <div className={styles.cardBody}>
            <div className={styles.row}>
              <span className={styles.label}>Layout</span>
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
            <div className={styles.row}>
              <span className={styles.label}>Columns</span>
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
            <div className={styles.row}>
              <span className={styles.label}>Image position</span>
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
            <div className={styles.row}>
              <span className={styles.label}>Image size</span>
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
            <div className={styles.row}>
              <span className={styles.label}>Card radius</span>
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
            <Toggle
              label="Show description"
              checked={draftConfig.items.showDescription}
              onChange={(v) => patchItems({ showDescription: v })}
            />
            <Toggle
              label="Show badges"
              checked={draftConfig.items.showBadges}
              onChange={(v) => patchItems({ showBadges: v })}
            />
          </div>
        </div>
      </div>

      {/* ── Right: Live preview ─────────────────────────────────────────────── */}
      <div className={styles.preview}>
        <span className={styles.previewLabel}>Live Preview</span>
        <div className={styles.phoneFrame}>
          <div className={styles.phoneNotch} />
          <div className={styles.phoneInner}>
            <div className={styles.previewScaler}>
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
