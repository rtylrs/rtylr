import React, { useState, useRef, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { motion, AnimatePresence } from 'motion/react';
import { Plus, ChevronDown, X } from 'lucide-react';

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
import { useLayoutPresetsStore } from '../../shared/stores/layoutPresetsStore';
import { useUserStore } from '../../shared/stores/userStore';
import { useMenuStore } from '../../shared/stores/menuStore';
import { useUpdateBusinessConfig } from '../../shared/hooks/queries/useBusiness';

// -- Placeholder image helper ------------------------------------------------

const PH = (w: number, h: number, bg: string, text: string) =>
  `https://placehold.co/${w}x${h}/${bg}/ffffff?text=${encodeURIComponent(text)}`;

// name and description are supplied by the component, because a module-level
// constant is evaluated before i18next has a language. Everything else here is
// a placeholder image or an empty field and has nothing to translate.
const DEFAULT_BUSINESS: Omit<BusinessInfo, 'name' | 'description'> = {
  coverImage: PH(800, 300, '1a1a1a', 'Cover'),
  logo: PH(120, 120, 'c8a96e', 'Logo'),
  address: '',
  phone: '',
  hours: '',
  averageRating: 0,
  reviewCount: 0,
};

const DEFAULT_CATEGORIES: MenuCategoryData[] = [];

// ─── Segment Button Group ──────────────────────────────────────────────────────

function SegmentGroup<T extends string>({
  value,
  options,
  onChange,
}: {
  value: T;
  options: { label: string; value: T }[];
  onChange: (v: T) => void;
}) {
  return (
    <div
      style={{
        display: 'flex',
        gap: '3px',
        background: 'var(--dash-input-bg)',
        border: '1px solid var(--dash-border)',
        borderRadius: '7px',
        padding: '3px',
      }}
    >
      {options.map((o) => (
        <button
          type="button"
          key={o.value}
          onClick={() => onChange(o.value)}
          style={{
            padding: '4px 10px',
            borderRadius: '4px',
            fontSize: '12px',
            fontWeight: 500,
            border: 'none',
            cursor: 'pointer',
            transition: 'all 0.15s ease',
            background: value === o.value ? '#4583C4' : 'transparent',
            color: value === o.value ? '#fff' : 'var(--dash-text-muted)',
            whiteSpace: 'nowrap',
          }}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}

// ─── Toggle Switch ─────────────────────────────────────────────────────────────

function ToggleRow({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '6px 0',
      }}
    >
      <span style={{ fontSize: '13px', color: 'var(--dash-text)' }}>{label}</span>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={() => onChange(!checked)}
        style={{
          width: 36,
          height: 20,
          borderRadius: 999,
          border: 'none',
          cursor: 'pointer',
          padding: 2,
          background: checked ? '#4583C4' : 'var(--dash-active)',
          transition: 'background 0.2s ease',
          position: 'relative',
          flexShrink: 0,
        }}
      >
        <span
          style={{
            display: 'block',
            width: 16,
            height: 16,
            borderRadius: '50%',
            background: '#fff',
            transform: checked ? 'translateX(16px)' : 'translateX(0)',
            transition: 'transform 0.2s ease',
          }}
        />
      </button>
    </div>
  );
}

// ─── Collapsible Section Card ──────────────────────────────────────────────────

function SectionCard({
  title,
  children,
  defaultOpen = true,
  index = 0,
}: {
  title: string;
  children: React.ReactNode;
  defaultOpen?: boolean;
  index?: number;
}) {
  const [open, setOpen] = useState(defaultOpen);

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2, delay: index * 0.04 }}
      style={{
        background: 'var(--dash-card)',
        border: '1px solid var(--dash-border)',
        borderRadius: '8px',
        overflow: 'hidden',
      }}
    >
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        style={{
          width: '100%',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '12px 16px',
          background: 'transparent',
          border: 'none',
          cursor: 'pointer',
          borderBottom: open ? '1px solid var(--dash-border)' : 'none',
          transition: 'border-color 0.15s',
        }}
      >
        <span
          style={{
            fontSize: '11px',
            fontWeight: 600,
            textTransform: 'uppercase',
            letterSpacing: 0,
            color: 'var(--dash-text-dim)',
          }}
        >
          {title}
        </span>
        <ChevronDown
          size={14}
          style={{
            color: 'var(--dash-text-dim)',
            transform: open ? 'rotate(180deg)' : 'rotate(0deg)',
            transition: 'transform 0.2s ease',
          }}
        />
      </button>

      <AnimatePresence initial={false}>
        {open && (
          <motion.div
            key="content"
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2, ease: 'easeInOut' }}
            style={{ overflow: 'hidden' }}
          >
            <div
              style={{
                padding: '12px 16px',
                display: 'flex',
                flexDirection: 'column',
                gap: '10px',
              }}
            >
              {children}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
}

// ─── Row with label + control ──────────────────────────────────────────────────

function ControlRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div
      style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12 }}
    >
      <span style={{ fontSize: '13px', color: 'var(--dash-text-muted)', flexShrink: 0 }}>
        {label}
      </span>
      {children}
    </div>
  );
}

// ─── Color picker row ─────────────────────────────────────────────────────────

function ColorRow({
  label,
  value,
  onChange,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <div
      style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12 }}
    >
      <span style={{ fontSize: '13px', color: 'var(--dash-text-muted)' }}>{label}</span>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <span style={{ fontSize: '12px', color: 'var(--dash-text-dim)', fontFamily: 'monospace' }}>
          {value}
        </span>
        <label
          style={{
            width: 28,
            height: 28,
            borderRadius: 6,
            overflow: 'hidden',
            border: '2px solid var(--dash-border)',
            cursor: 'pointer',
            background: value,
            display: 'block',
            flexShrink: 0,
          }}
        >
          <input
            type="color"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            style={{ opacity: 0, width: 1, height: 1, position: 'absolute' }}
          />
        </label>
      </div>
    </div>
  );
}

// ─── Phone Frame ──────────────────────────────────────────────────────────────

function PhoneFrame({
  config,
  businessInfo,
  menuCategories,
}: {
  config: MenuLayoutConfig;
  businessInfo: BusinessInfo;
  menuCategories: MenuCategoryData[];
}) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [scale, setScale] = useState(0.5);

  // Phone intrinsic dimensions
  const PHONE_W = 390;
  const PHONE_H = 844;

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const update = () => {
      const { height } = el.getBoundingClientRect();
      const availH = height - 64; // vertical padding
      const availW = el.getBoundingClientRect().width - 48;
      const byH = availH / PHONE_H;
      const byW = availW / PHONE_W;
      setScale(Math.min(byH, byW, 0.72));
    };
    update();
    const ro = new ResizeObserver(update);
    ro.observe(el);
    return () => ro.disconnect();
  }, []);

  return (
    <div
      ref={containerRef}
      style={{
        flex: 1,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        width: '100%',
        height: '100%',
      }}
    >
      {/* Outer phone shell */}
      <div
        style={{
          width: PHONE_W * scale,
          height: PHONE_H * scale,
          borderRadius: 44 * scale,
          background: '#18181b',
          boxShadow:
            '0 0 0 1px rgba(255,255,255,0.08), 0 32px 80px rgba(0,0,0,0.7), 0 8px 24px rgba(0,0,0,0.5)',
          position: 'relative',
          overflow: 'hidden',
          flexShrink: 0,
        }}
      >
        {/* Side button indicators (decorative) */}
        <div
          style={{
            position: 'absolute',
            left: -3 * scale,
            top: 120 * scale,
            width: 3 * scale,
            height: 32 * scale,
            background: '#2a2a2e',
            borderRadius: '2px 0 0 2px',
          }}
        />
        <div
          style={{
            position: 'absolute',
            left: -3 * scale,
            top: 162 * scale,
            width: 3 * scale,
            height: 64 * scale,
            background: '#2a2a2e',
            borderRadius: '2px 0 0 2px',
          }}
        />
        <div
          style={{
            position: 'absolute',
            left: -3 * scale,
            top: 236 * scale,
            width: 3 * scale,
            height: 64 * scale,
            background: '#2a2a2e',
            borderRadius: '2px 0 0 2px',
          }}
        />
        <div
          style={{
            position: 'absolute',
            right: -3 * scale,
            top: 160 * scale,
            width: 3 * scale,
            height: 80 * scale,
            background: '#2a2a2e',
            borderRadius: '0 2px 2px 0',
          }}
        />

        {/* Screen bezel */}
        <div
          style={{
            position: 'absolute',
            inset: 6 * scale,
            borderRadius: (44 - 6) * scale,
            background: '#fff',
            overflow: 'hidden',
          }}
        >
          {/* Dynamic island */}
          <div
            style={{
              position: 'absolute',
              top: 10 * scale,
              left: '50%',
              transform: 'translateX(-50%)',
              width: 120 * scale,
              height: 34 * scale,
              borderRadius: 20 * scale,
              background: '#18181b',
              zIndex: 20,
            }}
          />

          {/* Status bar */}
          <div
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              height: 54 * scale,
              background: 'rgba(0,0,0,0.02)',
              zIndex: 15,
              display: 'flex',
              alignItems: 'flex-end',
              justifyContent: 'space-between',
              padding: `0 ${16 * scale}px ${4 * scale}px`,
              fontSize: 10 * scale,
              fontWeight: 600,
              color: '#000',
            }}
          >
            <span>9:41</span>
            <span style={{ display: 'flex', gap: 4 * scale, alignItems: 'center' }}>
              <span>●●●</span>
              <span>WiFi</span>
              <span>⬛</span>
            </span>
          </div>

          {/* Scrollable menu content */}
          <div
            style={{
              position: 'absolute',
              inset: 0,
              top: 54 * scale,
              overflowY: 'auto',
              overflowX: 'hidden',
              WebkitOverflowScrolling: 'touch',
            }}
          >
            <div
              style={{
                transform: `scale(${scale})`,
                transformOrigin: 'top left',
                width: `${100 / scale}%`,
              }}
            >
              <MenuLayout business={businessInfo} categories={menuCategories} config={config} />
            </div>
          </div>

          {/* Home indicator */}
          <div
            style={{
              position: 'absolute',
              bottom: 8 * scale,
              left: '50%',
              transform: 'translateX(-50%)',
              width: 130 * scale,
              height: 5 * scale,
              borderRadius: 999,
              background: 'rgba(0,0,0,0.2)',
              zIndex: 20,
            }}
          />
        </div>
      </div>
    </div>
  );
}

// ─── Main Component ────────────────────────────────────────────────────────────

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

  const { t } = useTranslation();
  const { user } = useUserStore();
  const { items: menuItems } = useMenuStore();

  const defaultBusiness: BusinessInfo = {
    ...DEFAULT_BUSINESS,
    name: t('design.businessPlaceholder.name'),
    description: t('design.businessPlaceholder.description'),
  };

  // Build business info and menu categories from real data
  const u = user as Record<string, unknown> | null;
  const businessInfo: BusinessInfo = u
    ? {
        name: (u.businessName as string) || defaultBusiness.name,
        description: (u.businessDescription as string) || defaultBusiness.description,
        coverImage: (u.coverImage as string) || DEFAULT_BUSINESS.coverImage,
        logo: (u.logo as string) || DEFAULT_BUSINESS.logo,
        address: (u.address as string) || DEFAULT_BUSINESS.address,
        phone: (u.phone as string) || DEFAULT_BUSINESS.phone,
        hours: (u.hours as string) || DEFAULT_BUSINESS.hours,
        averageRating: (u.averageRating as number) ?? DEFAULT_BUSINESS.averageRating,
        reviewCount: (u.reviewCount as number) ?? DEFAULT_BUSINESS.reviewCount,
      }
    : defaultBusiness;

  interface RawMenuItem {
    _id?: string;
    name?: string;
    description?: string;
    price?: number;
    currency?: string;
    available?: boolean;
    image?: string;
    badges?: string[];
  }
  interface RawCategory {
    _id?: string;
    name?: string;
    type?: string;
    menuItems?: RawMenuItem[];
  }
  const menuCategories: MenuCategoryData[] =
    menuItems && menuItems.length > 0
      ? ((menuItems as RawCategory[])
          .filter((item) => item.type === 'category')
          .map((cat) => ({
            id: cat._id,
            name: cat.name,
            items: (cat.menuItems || []).map((mi) => ({
              id: mi._id,
              name: mi.name,
              description: mi.description || '',
              price: mi.price ?? 0,
              currency: mi.currency || 'USD',
              available: mi.available !== false,
              image: mi.image || '',
              badges: mi.badges || [],
            })),
          })) as MenuCategoryData[])
      : DEFAULT_CATEGORIES;

  const [newPresetName, setNewPresetName] = useState('');
  const [saved, setSaved] = useState(false);

  const [fontFamily, setFontFamily] = useState('system');
  const [fontSize, setFontSize] = useState('md');
  const colorPrimary = draftConfig.colors?.primary ?? '#4583C4';
  const colorAccent = draftConfig.colors?.accent ?? '#f59e0b';
  const colorBg = draftConfig.colors?.bg ?? '#ffffff';

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

  const updateConfig = useUpdateBusinessConfig();

  const handleSave = () => {
    saveActivePreset();
    const { presets, activePresetId, draftConfig } = useLayoutPresetsStore.getState();
    updateConfig.mutate({
      layout_presets: presets,
      active_preset_id: activePresetId,
      draft_config: draftConfig,
    });
    setSaved(true);
    setTimeout(() => setSaved(false), 1800);
  };

  return (
    <>
      <HelmetContainer title={`${t('design.pageTitle')} · rtylr`} />
      <div
        style={{
          display: 'grid',
          gridTemplateRows: 'auto 1fr',
          // Escape DefaultLayout's p-6 padding to fill edge-to-edge
          margin: '-24px',
          height: 'calc(100% + 48px)',
          overflow: 'hidden',
          background: 'var(--dash-bg)',
        }}
      >
        {/* ── Header bar ─────────────────────────────────────────────────────── */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'flex-end',
            padding: '0 16px',
            height: 52,
            borderBottom: '1px solid var(--dash-border)',
            background: 'var(--dash-surface)',
            flexShrink: 0,
            gap: 12,
          }}
        >
          <div style={{ flex: 1, minWidth: 0 }}>
            <h1 style={{ margin: 0, fontSize: 15, fontWeight: 650, color: 'var(--dash-text)' }}>
              {t('design.pageTitle')}
            </h1>
            <p style={{ margin: '2px 0 0', fontSize: 12, color: 'var(--dash-text-dim)' }}>
              {t('design.subtitle')}
            </p>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            {activePreset && (
              <span style={{ fontSize: 12, color: 'var(--dash-text-dim)' }}>
                {activePreset.name}
              </span>
            )}
            <motion.button
              whileTap={{ scale: 0.96 }}
              onClick={handleSave}
              style={{
                padding: '6px 14px',
                borderRadius: 7,
                background: saved ? 'var(--accent-muted)' : 'var(--accent)',
                color: saved ? 'var(--accent-on-muted)' : 'var(--dash-on-accent)',
                fontSize: 13,
                fontWeight: 600,
                border: saved ? '1px solid rgba(69,131,196,0.3)' : 'none',
                cursor: 'pointer',
                transition: 'all 0.2s ease',
              }}
            >
              {saved ? 'Saved' : 'Save'}
            </motion.button>
          </div>
        </div>

        {/* ── Body ───────────────────────────────────────────────────────────── */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '220px minmax(300px, 340px) minmax(0, 1fr)',
            overflow: 'hidden',
            height: '100%',
          }}
        >
          {/* ── Presets sidebar ──────────────────────────────────────────────── */}
          <aside
            style={{
              display: 'flex',
              flexDirection: 'column',
              height: '100%',
              overflow: 'hidden',
              borderRight: '1px solid var(--dash-border)',
              background: 'var(--dash-surface)',
            }}
          >
            <div
              style={{
                padding: '12px 14px 8px',
                borderBottom: '1px solid var(--dash-border)',
                flexShrink: 0,
              }}
            >
              <span
                style={{
                  fontSize: 10,
                  fontWeight: 600,
                  textTransform: 'uppercase',
                  letterSpacing: 0,
                  color: 'var(--dash-text-dim)',
                }}
              >
                {t('design.layouts')}
              </span>
            </div>

            <div style={{ flex: 1, overflowY: 'auto', padding: '6px' }}>
              {presets.map((preset) => (
                <div
                  key={preset.id}
                  onClick={() => activatePreset(preset.id)}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 8,
                    padding: '8px 10px',
                    borderRadius: 8,
                    cursor: 'pointer',
                    background: preset.id === activePresetId ? 'var(--dash-active)' : 'transparent',
                    marginBottom: 2,
                    transition: 'background 0.15s',
                  }}
                >
                  {preset.id === activePresetId && (
                    <span
                      style={{
                        width: 6,
                        height: 6,
                        borderRadius: '50%',
                        background: '#4583C4',
                        flexShrink: 0,
                      }}
                    />
                  )}
                  <span
                    style={{
                      fontSize: 13,
                      fontWeight: 500,
                      color:
                        preset.id === activePresetId
                          ? 'var(--dash-text)'
                          : 'var(--dash-text-muted)',
                      flex: 1,
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                    }}
                  >
                    {preset.name}
                  </span>
                  {presets.length > 1 && (
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        deletePreset(preset.id);
                      }}
                      style={{
                        background: 'none',
                        border: 'none',
                        cursor: 'pointer',
                        padding: 2,
                        borderRadius: 4,
                        color: 'var(--dash-text-dim)',
                        display: 'flex',
                        alignItems: 'center',
                        opacity: 0.6,
                        transition: 'opacity 0.15s',
                      }}
                    >
                      <X size={12} />
                    </button>
                  )}
                </div>
              ))}
            </div>

            <div
              style={{
                padding: '8px',
                borderTop: '1px solid var(--dash-border)',
                display: 'flex',
                gap: 6,
                flexShrink: 0,
              }}
            >
              <input
                placeholder={t('design.newLayoutPlaceholder')}
                value={newPresetName}
                onChange={(e) => setNewPresetName(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleAddPreset()}
                style={{
                  flex: 1,
                  background: 'var(--dash-input-bg)',
                  border: '1px solid var(--dash-border)',
                  borderRadius: 7,
                  padding: '6px 10px',
                  fontSize: 12,
                  color: 'var(--dash-text)',
                  outline: 'none',
                }}
              />
              <button
                type="button"
                onClick={handleAddPreset}
                disabled={!newPresetName.trim()}
                style={{
                  padding: '6px 8px',
                  background: newPresetName.trim() ? '#4583C4' : 'var(--dash-input-bg)',
                  border: '1px solid var(--dash-border)',
                  borderRadius: 7,
                  color: newPresetName.trim() ? '#fff' : 'var(--dash-text-dim)',
                  cursor: newPresetName.trim() ? 'pointer' : 'default',
                  display: 'flex',
                  alignItems: 'center',
                  transition: 'all 0.15s',
                }}
              >
                <Plus size={14} />
              </button>
            </div>
          </aside>

          {/* ── Controls panel ───────────────────────────────────────────────── */}
          <div
            style={{
              overflowY: 'auto',
              padding: '16px 14px',
              display: 'flex',
              flexDirection: 'column',
              gap: 10,
              borderRight: '1px solid var(--dash-border)',
            }}
          >
            {/* Pages */}
            <SectionCard title={t('design.sections.pages')} index={0}>
              {(
                ['about', 'gallery', 'reviews', 'offers'] as (keyof MenuLayoutConfig['pages'])[]
              ).map((key) => (
                <ToggleRow
                  key={key}
                  label={t(`design.pages.${key}`)}
                  checked={draftConfig.pages[key] === true}
                  onChange={(v) => patchPages({ [key]: v })}
                />
              ))}
            </SectionCard>

            {/* Header */}
            <SectionCard title={t('design.sections.header')} index={1}>
              <ControlRow label={t('design.header.coverHeight')}>
                <SegmentGroup<CoverHeight>
                  value={draftConfig.header.coverHeight}
                  options={[
                    { label: t('design.none'), value: 'none' },
                    { label: 'S', value: 'sm' },
                    { label: 'M', value: 'md' },
                    { label: 'L', value: 'lg' },
                  ]}
                  onChange={(v) => patchHeader({ coverHeight: v })}
                />
              </ControlRow>
              <ToggleRow
                label={t('design.header.sticky')}
                checked={draftConfig.header.sticky}
                onChange={(v) => patchHeader({ sticky: v })}
              />
              <ToggleRow
                label={t('design.header.showLogo')}
                checked={draftConfig.header.showLogo}
                onChange={(v) => patchHeader({ showLogo: v })}
              />
              <ToggleRow
                label={t('design.header.showDescription')}
                checked={draftConfig.header.showDescription}
                onChange={(v) => patchHeader({ showDescription: v })}
              />
              <ToggleRow
                label={t('design.header.showMeta')}
                checked={draftConfig.header.showMeta}
                onChange={(v) => patchHeader({ showMeta: v })}
              />
            </SectionCard>

            {/* Nav Tabs */}
            <SectionCard title={t('design.sections.nav')} index={2}>
              <ControlRow label={t('design.nav.style')}>
                <SegmentGroup<NavStyle>
                  value={draftConfig.nav.style}
                  options={[
                    { label: t('design.nav.styleUnderline'), value: 'underline' },
                    { label: t('design.nav.stylePill'), value: 'pill' },
                    { label: t('design.nav.styleFlat'), value: 'flat' },
                  ]}
                  onChange={(v) => patchNav({ style: v })}
                />
              </ControlRow>
              <ControlRow label={t('design.nav.size')}>
                <SegmentGroup<NavSize>
                  value={draftConfig.nav.size}
                  options={[
                    { label: 'S', value: 'sm' },
                    { label: 'M', value: 'md' },
                    { label: 'L', value: 'lg' },
                  ]}
                  onChange={(v) => patchNav({ size: v })}
                />
              </ControlRow>
            </SectionCard>

            {/* Item Cards */}
            <SectionCard title={t('design.sections.items')} index={3}>
              <ControlRow label={t('design.items.layout')}>
                <SegmentGroup<ItemLayout>
                  value={draftConfig.items.layout}
                  options={[
                    { label: t('design.items.layoutGrid'), value: 'grid' },
                    { label: t('design.items.layoutList'), value: 'list' },
                    { label: t('design.items.layoutMinimal'), value: 'minimal' },
                  ]}
                  onChange={(v) => patchItems({ layout: v })}
                />
              </ControlRow>
              <ControlRow label={t('design.items.columns')}>
                <SegmentGroup<string>
                  value={String(draftConfig.items.columns)}
                  options={[
                    { label: '2', value: '2' },
                    { label: '3', value: '3' },
                    { label: '4', value: '4' },
                  ]}
                  onChange={(v) => patchItems({ columns: Number(v) as 2 | 3 | 4 })}
                />
              </ControlRow>
              <ControlRow label={t('design.items.imagePosition')}>
                <SegmentGroup<ImagePosition>
                  value={draftConfig.items.imagePosition}
                  options={[
                    { label: t('design.items.imageTop'), value: 'top' },
                    { label: t('design.items.imageRight'), value: 'right' },
                    { label: t('design.items.imageLeft'), value: 'left' },
                    { label: t('design.none'), value: 'none' },
                  ]}
                  onChange={(v) => patchItems({ imagePosition: v })}
                />
              </ControlRow>
              <ControlRow label={t('design.items.imageSize')}>
                <SegmentGroup<ImageSize>
                  value={draftConfig.items.imageSize}
                  options={[
                    { label: 'S', value: 'sm' },
                    { label: 'M', value: 'md' },
                    { label: 'L', value: 'lg' },
                  ]}
                  onChange={(v) => patchItems({ imageSize: v })}
                />
              </ControlRow>
              <ControlRow label={t('design.items.cardRadius')}>
                <SegmentGroup<CardRadius>
                  value={draftConfig.items.cardRadius}
                  options={[
                    { label: t('design.none'), value: 'none' },
                    { label: 'S', value: 'sm' },
                    { label: 'M', value: 'md' },
                    { label: 'L', value: 'lg' },
                  ]}
                  onChange={(v) => patchItems({ cardRadius: v })}
                />
              </ControlRow>
              <ToggleRow
                label={t('design.items.showDescription')}
                checked={draftConfig.items.showDescription}
                onChange={(v) => patchItems({ showDescription: v })}
              />
              <ToggleRow
                label={t('design.items.showBadges')}
                checked={draftConfig.items.showBadges}
                onChange={(v) => patchItems({ showBadges: v })}
              />
            </SectionCard>

            {/* Colors */}
            <SectionCard title={t('design.sections.colors')} index={4}>
              <ColorRow
                label={t('design.colors.primary')}
                value={colorPrimary}
                onChange={(v) => {
                  patchDraftConfig({
                    colors: { ...draftConfig.colors, primary: v, accent: colorAccent, bg: colorBg },
                  });
                }}
              />
              <ColorRow
                label={t('design.colors.accent')}
                value={colorAccent}
                onChange={(v) => {
                  patchDraftConfig({
                    colors: {
                      primary: colorPrimary,
                      ...draftConfig.colors,
                      accent: v,
                      bg: colorBg,
                    },
                  });
                }}
              />
              <ColorRow
                label={t('design.colors.background')}
                value={colorBg}
                onChange={(v) => {
                  patchDraftConfig({
                    colors: {
                      primary: colorPrimary,
                      accent: colorAccent,
                      ...draftConfig.colors,
                      bg: v,
                    },
                  });
                }}
              />
            </SectionCard>

            {/* Typography */}
            <SectionCard title={t('design.sections.typography')} index={5}>
              <ControlRow label={t('design.typography.fontFamily')}>
                <select
                  value={fontFamily}
                  onChange={(e) => setFontFamily(e.target.value)}
                  style={{
                    background: 'var(--dash-input-bg)',
                    border: '1px solid var(--dash-border)',
                    borderRadius: 7,
                    padding: '5px 8px',
                    fontSize: 12,
                    color: 'var(--dash-text)',
                    outline: 'none',
                    cursor: 'pointer',
                  }}
                >
                  <option value="system">System UI</option>
                  <option value="inter">Inter</option>
                  <option value="jakarta">Plus Jakarta Sans</option>
                  <option value="merriweather">Merriweather</option>
                  <option value="poppins">Poppins</option>
                </select>
              </ControlRow>
              <ControlRow label={t('design.typography.fontSize')}>
                <SegmentGroup<string>
                  value={fontSize}
                  options={[
                    { label: 'S', value: 'sm' },
                    { label: 'M', value: 'md' },
                    { label: 'L', value: 'lg' },
                  ]}
                  onChange={setFontSize}
                />
              </ControlRow>
            </SectionCard>
          </div>

          {/* ── Phone Preview ─────────────────────────────────────────────────── */}
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              padding: '20px 0 12px',
              background: 'var(--dash-bg)',
              overflow: 'hidden',
              height: '100%',
            }}
          >
            <span
              style={{
                fontSize: 10,
                fontWeight: 600,
                textTransform: 'uppercase',
                letterSpacing: 0,
                color: 'var(--dash-text-dim)',
                marginBottom: 16,
                flexShrink: 0,
              }}
            >
              {t('design.preview')}
            </span>
            <div
              style={{
                flex: 1,
                width: '100%',
                overflow: 'hidden',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <PhoneFrame
                config={draftConfig}
                businessInfo={businessInfo}
                menuCategories={menuCategories}
              />
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default Design;
