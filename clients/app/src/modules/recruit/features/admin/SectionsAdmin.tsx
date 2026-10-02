import AdminQueryState from '../../shared/components/AdminQueryState';
import React, { useEffect, useRef, useState } from 'react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { uploadAsset } from '@voxire/api';
import {
  Plus,
  GripVertical,
  ChevronDown,
  ChevronRight,
  Trash2,
  Eye,
  EyeOff,
  Sparkles,
  Users,
  Quote,
  PlayCircle,
  HelpCircle,
  Star,
  MousePointerClick,
  Upload,
  LayoutPanelTop,
} from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';

import {
  useAdminSections,
  useCreateSection,
  useUpdateSection,
  useDeleteSection,
  useReorderSections,
  type SectionBody,
} from '../../shared/hooks/queries/useAdmin';
import type { Section, SectionKind, PerkItem, FaqItem, ValueItem } from '../../shared/lib/types';

// KIND_META is keyed by a narrower set of known kinds than SectionKind
// (which accepts any string for forward-compat). Indexing with a
// SectionKind value works at runtime but TypeScript needs the lookup
// to widen to a plain-string record.
type KnownKind = 'perks' | 'culture' | 'testimonial' | 'video' | 'faq' | 'values' | 'cta';

const KIND_META: Record<KnownKind, { icon: React.ReactNode }> = {
  perks: { icon: <Sparkles className="h-4 w-4" /> },
  culture: { icon: <Users className="h-4 w-4" /> },
  testimonial: { icon: <Quote className="h-4 w-4" /> },
  video: { icon: <PlayCircle className="h-4 w-4" /> },
  faq: { icon: <HelpCircle className="h-4 w-4" /> },
  values: { icon: <Star className="h-4 w-4" /> },
  cta: { icon: <MousePointerClick className="h-4 w-4" /> },
};

// Exported so the locale suite can assert a label and a blurb exist for every
// kind. Both are looked up with a runtime built key.
export const ORDERED_KINDS: KnownKind[] = [
  'perks',
  'culture',
  'testimonial',
  'video',
  'faq',
  'values',
  'cta',
];

// lookupMeta safely reads from KIND_META with a runtime-typed string
// without tripping TypeScript's index-signature checks.
function lookupMeta(kind: string): (typeof KIND_META)[KnownKind] | undefined {
  return (KIND_META as Record<string, (typeof KIND_META)[KnownKind]>)[kind];
}

const SectionsAdmin: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation('recruitCommon');
  // A kind outside the known set has no copy to show, so it falls back to its
  // own value, which is what the English label did before.
  const kindLabel = (kind: string): string =>
    lookupMeta(kind) ? t(`recruitAdmin.sections.kind.${kind}.label`) : kind;
  const { data, isLoading, isError, refetch } = useAdminSections();
  const create = useCreateSection();
  const update = useUpdateSection();
  const del = useDeleteSection();
  const reorder = useReorderSections();

  // Local ordered copy so drag-drop is instant; server is the backstop.
  const [localOrder, setLocalOrder] = useState<Section[]>([]);
  useEffect(() => {
    setLocalOrder(data?.sections ?? []);
  }, [data?.sections]);

  const [expandedId, setExpandedId] = useState<number | null>(null);
  const [adding, setAdding] = useState(false);
  const [dragId, setDragId] = useState<number | null>(null);

  function commitReorder(next: Section[]) {
    setLocalOrder(next);
    reorder.mutate(
      next.map((s) => s.id),
      { onError: () => toast.error(t('recruitAdmin.sections.orderFailed')) }
    );
  }

  function onDragStart(id: number) {
    setDragId(id);
  }
  function onDragOver(e: React.DragEvent, overId: number) {
    e.preventDefault();
    if (dragId === null || dragId === overId) return;
    const fromIdx = localOrder.findIndex((s) => s.id === dragId);
    const toIdx = localOrder.findIndex((s) => s.id === overId);
    if (fromIdx < 0 || toIdx < 0) return;
    const next = [...localOrder];
    const [moved] = next.splice(fromIdx, 1);
    next.splice(toIdx, 0, moved);
    setLocalOrder(next);
  }
  function onDragEnd() {
    if (dragId !== null) commitReorder(localOrder);
    setDragId(null);
  }

  if (isError) return <AdminQueryState retry={() => void refetch()} />;

  return (
    <>
      <HelmetContainer title={t('recruitAdmin.sections.metaTitle')} />
      <div className="px-10 py-8">
        <div className="flex items-baseline justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">
              {t('recruitAdmin.sections.heading')}
            </h1>
            <p className="mt-1 text-sm text-[var(--dash-text-muted)]">
              {t('recruitAdmin.sections.hint')}
            </p>
          </div>
          <button
            type="button"
            onClick={() => setAdding(true)}
            className="inline-flex items-center gap-2 rounded-md bg-[var(--dash-hover)] px-4 py-2 text-sm font-semibold text-[var(--dash-text)] hover:bg-[var(--dash-active)]"
          >
            <Plus className="h-4 w-4" />
            {t('recruitAdmin.sections.addSection')}
          </button>
        </div>

        {adding && (
          <AddSectionPicker
            onCancel={() => setAdding(false)}
            onPick={(kind) => {
              const body: SectionBody = { kind, enabled: true };
              create.mutate(body, {
                onSuccess: (res) => {
                  toast.success(t('recruitAdmin.sections.added'));
                  setAdding(false);
                  if (res?.section) setExpandedId(res.section.id);
                },
                onError: () => toast.error(t('recruitAdmin.sections.addFailed')),
              });
            }}
          />
        )}

        <div
          className="mt-6 overflow-hidden rounded-lg border"
          style={{
            borderColor: 'var(--recruit-admin-border)',
            background: 'var(--recruit-admin-surface)',
          }}
        >
          {isLoading ? (
            <div className="p-10 text-center text-sm text-[var(--dash-text-muted)]">
              {t('recruitAdmin.sections.loading')}
            </div>
          ) : localOrder.length === 0 ? (
            <EmptyState
              icon={LayoutPanelTop}
              title={t('recruitCommon.emptyStates.sectionsNone')}
              action={{
                label: t('recruitCommon.emptyStates.addSection'),
                onClick: () => setAdding(true),
              }}
            />
          ) : (
            <ul>
              {localOrder.map((s) => (
                <li
                  key={s.id}
                  draggable
                  onDragStart={() => onDragStart(s.id)}
                  onDragOver={(e) => onDragOver(e, s.id)}
                  onDragEnd={onDragEnd}
                  className="border-b last:border-b-0"
                  style={{
                    borderColor: 'var(--recruit-admin-border)',
                    opacity: dragId === s.id ? 0.4 : 1,
                  }}
                >
                  <div className="flex items-center gap-3 px-4 py-3 hover:bg-[var(--dash-hover)]">
                    <GripVertical
                      className="h-4 w-4 cursor-grab text-[var(--dash-text-dim)] active:cursor-grabbing"
                      aria-hidden
                    />
                    <span className="text-[var(--dash-text-muted)]">
                      {lookupMeta(s.kind)?.icon}
                    </span>
                    <div className="min-w-0 flex-1">
                      <div className="truncate text-sm font-medium">
                        {s.title || kindLabel(s.kind)}
                      </div>
                      <div className="truncate text-xs text-[var(--dash-text-dim)]">
                        {kindLabel(s.kind)}
                        {!s.enabled && ' · hidden'}
                      </div>
                    </div>
                    <button
                      type="button"
                      onClick={() =>
                        update.mutate(
                          {
                            id: s.id,
                            kind: s.kind,
                            title: s.title,
                            body: s.body,
                            mediaUrl: s.mediaUrl,
                            itemsJson: s.itemsJson,
                            position: s.position,
                            enabled: !s.enabled,
                          },
                          { onError: () => toast.error(t('recruitAdmin.sections.toggleFailed')) }
                        )
                      }
                      title={
                        s.enabled
                          ? t('recruitAdmin.sections.hideFromPublic')
                          : t('recruitAdmin.sections.showOnPublic')
                      }
                      className="rounded-md p-1.5 text-[var(--dash-text-muted)] hover:bg-[var(--dash-hover)] hover:text-[var(--dash-text)]"
                    >
                      {s.enabled ? <Eye className="h-4 w-4" /> : <EyeOff className="h-4 w-4" />}
                    </button>
                    <button
                      type="button"
                      onClick={() => setExpandedId(expandedId === s.id ? null : s.id)}
                      className="rounded-md p-1.5 text-[var(--dash-text-muted)] hover:bg-[var(--dash-hover)] hover:text-[var(--dash-text)]"
                    >
                      {expandedId === s.id ? (
                        <ChevronDown className="h-4 w-4" />
                      ) : (
                        <ChevronRight className="h-4 w-4" />
                      )}
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        toast.warning(t('recruitAdmin.sections.confirmDelete'), {
                          action: {
                            label: t('recruitAdmin.sections.delete'),
                            onClick: () =>
                              del.mutate(s.id, {
                                onSuccess: () => toast.success(t('recruitAdmin.sections.deleted')),
                                onError: () => toast.error(t('recruitAdmin.sections.deleteFailed')),
                              }),
                          },
                        });
                      }}
                      className="rounded-md p-1.5 text-[var(--dash-text-muted)] hover:bg-[var(--dash-hover)] hover:text-red-400"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>

                  {expandedId === s.id && (
                    <div
                      className="border-t bg-[var(--dash-hover)] px-4 py-5"
                      style={{ borderColor: 'var(--recruit-admin-border)' }}
                    >
                      <SectionEditor
                        section={s}
                        onSave={(body) =>
                          update.mutate(body, {
                            onSuccess: () => toast.success(t('recruitAdmin.sections.saved')),
                            onError: () => toast.error(t('recruitAdmin.sections.saveFailed')),
                          })
                        }
                      />
                    </div>
                  )}
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </>
  );
};

// ── AddSectionPicker - grid of kinds to drop in ──────────────────────

const AddSectionPicker: React.FC<{
  onPick: (kind: SectionKind) => void;
  onCancel: () => void;
}> = ({ onPick, onCancel }) => {
  const { t } = useTranslation('recruitCommon');
  return (
    <div
      className="mt-6 rounded-lg border p-5"
      style={{
        borderColor: 'var(--recruit-admin-border)',
        background: 'var(--recruit-admin-surface)',
      }}
    >
      <div className="flex items-center justify-between">
        <div className="text-sm font-medium">{t('recruitAdmin.sections.pickKind')}</div>
        <button
          type="button"
          onClick={onCancel}
          className="rounded-md px-2 py-1 text-xs text-[var(--dash-text-muted)] hover:bg-[var(--dash-hover)]"
        >
          {t('recruitAdmin.sections.cancel')}
        </button>
      </div>
      <div className="mt-4 grid grid-cols-2 gap-2 sm:grid-cols-4">
        {ORDERED_KINDS.map((k) => {
          const meta = KIND_META[k];
          return (
            <button
              type="button"
              key={k}
              onClick={() => onPick(k)}
              className="flex flex-col items-start rounded-md border p-3 text-left transition hover:-translate-y-0.5 hover:border-white/30"
              style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
            >
              <span className="mb-2 text-[var(--dash-text-muted)]">{meta.icon}</span>
              <span className="text-sm font-medium">
                {t(`recruitAdmin.sections.kind.${k}.label`)}
              </span>
              <span className="mt-0.5 text-[10px] text-[var(--dash-text-dim)]">
                {t(`recruitAdmin.sections.kind.${k}.blurb`)}
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
};

// ── SectionEditor - per-kind forms ───────────────────────────────────

const SectionEditor: React.FC<{
  section: Section;
  onSave: (body: SectionBody) => void;
}> = ({ section, onSave }) => {
  const { t } = useTranslation('recruitCommon');
  const [title, setTitle] = useState(section.title);
  const [body, setBody] = useState(section.body);
  const [mediaUrl, setMediaUrl] = useState(section.mediaUrl);
  const [uploadingMedia, setUploadingMedia] = useState(false);
  const mediaInputRef = useRef<HTMLInputElement>(null);
  const [items, setItems] = useState<unknown[]>(() => {
    if (!section.itemsJson) return [];
    try {
      const parsed = JSON.parse(section.itemsJson);
      return Array.isArray(parsed) ? parsed : [];
    } catch {
      return [];
    }
  });

  useEffect(() => {
    setTitle(section.title);
    setBody(section.body);
    setMediaUrl(section.mediaUrl);
    try {
      const parsed = section.itemsJson ? JSON.parse(section.itemsJson) : [];
      setItems(Array.isArray(parsed) ? parsed : []);
    } catch {
      setItems([]);
    }
    // Sync-on-select: re-running when the section fields change would wipe
    // in-progress edits, so this intentionally re-syncs on id change alone.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [section.id]);

  function save() {
    onSave({
      id: section.id,
      kind: section.kind,
      title,
      body,
      mediaUrl,
      itemsJson: items.length > 0 ? JSON.stringify(items) : '',
      position: section.position,
      enabled: section.enabled,
    });
  }

  async function uploadMedia(file: File | undefined | null) {
    if (!file) return;
    setUploadingMedia(true);
    try {
      const uploaded = await uploadAsset(file);
      setMediaUrl(uploaded.url);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('recruitAdmin.sections.uploadFailed'));
    } finally {
      setUploadingMedia(false);
      if (mediaInputRef.current) mediaInputRef.current.value = '';
    }
  }

  return (
    <div className="space-y-4">
      <Field label={t('recruitAdmin.sections.sectionTitle')}>
        <Input
          value={title}
          onChange={setTitle}
          placeholder={t('recruitAdmin.sections.sectionTitlePlaceholder')}
        />
      </Field>

      {(section.kind === 'culture' || section.kind === 'cta' || !lookupMeta(section.kind)) && (
        <Field label={t('recruitAdmin.sections.body')}>
          <Textarea value={body} onChange={setBody} rows={4} />
        </Field>
      )}

      {section.kind === 'testimonial' && (
        <>
          <Field label={t('recruitAdmin.sections.quote')}>
            <Textarea value={body} onChange={setBody} rows={3} />
          </Field>
          <Field
            label={t('recruitAdmin.sections.attribution')}
            hint={t('recruitAdmin.sections.attributionHint')}
          >
            <Input
              value={mediaUrl}
              onChange={setMediaUrl}
              placeholder={t('recruitAdmin.sections.attributionPlaceholder')}
            />
          </Field>
        </>
      )}

      {section.kind === 'culture' && (
        <Field
          label={t('recruitAdmin.sections.photoUrl')}
          hint={t('recruitAdmin.sections.photoHint')}
        >
          <div className="flex gap-2">
            <Input value={mediaUrl} onChange={setMediaUrl} type="url" />
            <button
              type="button"
              onClick={() => mediaInputRef.current?.click()}
              disabled={uploadingMedia}
              className="inline-flex items-center gap-2 rounded-md border border-[var(--dash-border)] px-3 text-sm font-semibold text-[var(--dash-text)]/80 hover:bg-[var(--dash-hover)] disabled:opacity-60"
            >
              <Upload size={14} />
              {uploadingMedia ? 'Uploading...' : 'Upload'}
            </button>
            <input
              ref={mediaInputRef}
              type="file"
              accept="image/*"
              className="hidden"
              onChange={(e) => {
                void uploadMedia(e.target.files?.[0]);
              }}
            />
          </div>
        </Field>
      )}

      {section.kind === 'video' && (
        <Field
          label={t('recruitAdmin.sections.videoUrl')}
          hint={t('recruitAdmin.sections.videoHint')}
        >
          <div className="flex gap-2">
            <Input value={mediaUrl} onChange={setMediaUrl} type="url" />
            <button
              type="button"
              onClick={() => mediaInputRef.current?.click()}
              disabled={uploadingMedia}
              className="inline-flex items-center gap-2 rounded-md border border-[var(--dash-border)] px-3 text-sm font-semibold text-[var(--dash-text)]/80 hover:bg-[var(--dash-hover)] disabled:opacity-60"
            >
              <Upload size={14} />
              {uploadingMedia ? 'Uploading...' : 'Upload'}
            </button>
            <input
              ref={mediaInputRef}
              type="file"
              accept="video/*"
              className="hidden"
              onChange={(e) => {
                void uploadMedia(e.target.files?.[0]);
              }}
            />
          </div>
        </Field>
      )}

      {section.kind === 'cta' && (
        <Field label={t('recruitAdmin.sections.buttonUrl')}>
          <Input value={mediaUrl} onChange={setMediaUrl} type="url" placeholder="https://…" />
        </Field>
      )}

      {section.kind === 'perks' && (
        <PerksEditor items={(items as PerkItem[]) ?? []} onChange={(next) => setItems(next)} />
      )}
      {section.kind === 'faq' && (
        <FaqEditor items={(items as FaqItem[]) ?? []} onChange={(next) => setItems(next)} />
      )}
      {section.kind === 'values' && (
        <ValuesEditor items={(items as ValueItem[]) ?? []} onChange={(next) => setItems(next)} />
      )}

      <div className="flex justify-end">
        <button
          type="button"
          onClick={save}
          className="rounded-md bg-[var(--dash-hover)] px-4 py-2 text-sm font-semibold text-[var(--dash-text)] hover:bg-[var(--dash-active)]"
        >
          {t('recruitAdmin.sections.saveSection')}
        </button>
      </div>
    </div>
  );
};

// ── item editors (perks, faq, values) ────────────────────────────────

const PerksEditor: React.FC<{
  items: PerkItem[];
  onChange: (items: PerkItem[]) => void;
}> = ({ items, onChange }) => {
  const { t } = useTranslation('recruitCommon');
  return (
    <div>
      <div className="mb-2 text-xs font-medium uppercase tracking-wider text-[var(--dash-text-muted)]">
        {t('recruitAdmin.sections.perks')}
      </div>
      <div className="space-y-2">
        {items.map((p, i) => (
          <div key={i} className="flex gap-2">
            <Input
              value={p.icon ?? ''}
              onChange={(v) => onChange(items.map((x, j) => (i === j ? { ...x, icon: v } : x)))}
              placeholder={t('recruitAdmin.sections.perkIconPlaceholder')}
            />
            <Input
              value={p.label}
              onChange={(v) => onChange(items.map((x, j) => (i === j ? { ...x, label: v } : x)))}
              placeholder={t('recruitAdmin.sections.perkLabelPlaceholder')}
            />
            <Input
              value={p.blurb ?? ''}
              onChange={(v) => onChange(items.map((x, j) => (i === j ? { ...x, blurb: v } : x)))}
              placeholder={t('recruitAdmin.sections.perkBlurbPlaceholder')}
            />
            <button
              type="button"
              onClick={() => onChange(items.filter((_, j) => j !== i))}
              className="rounded-md p-2 text-[var(--dash-text-muted)] hover:bg-[var(--dash-hover)] hover:text-red-400"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          </div>
        ))}
        <button
          type="button"
          onClick={() => onChange([...items, { label: '', icon: 'Sparkles' }])}
          className="inline-flex items-center gap-1.5 rounded-md border border-dashed border-[var(--dash-border)] px-3 py-2 text-xs text-[var(--dash-text-muted)] hover:bg-[var(--dash-hover)]"
        >
          <Plus className="h-3 w-3" />
          {t('recruitAdmin.sections.addPerk')}
        </button>
      </div>
    </div>
  );
};

const FaqEditor: React.FC<{
  items: FaqItem[];
  onChange: (items: FaqItem[]) => void;
}> = ({ items, onChange }) => {
  const { t } = useTranslation('recruitCommon');
  return (
    <div>
      <div className="mb-2 text-xs font-medium uppercase tracking-wider text-[var(--dash-text-muted)]">
        {t('recruitAdmin.sections.faqTitle')}
      </div>
      <div className="space-y-2">
        {items.map((f, i) => (
          <div
            key={i}
            className="rounded-md border p-3"
            style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
          >
            <Input
              value={f.question}
              onChange={(v) => onChange(items.map((x, j) => (i === j ? { ...x, question: v } : x)))}
              placeholder={t('recruitAdmin.sections.faqQuestionPlaceholder')}
            />
            <div className="mt-2">
              <Textarea
                value={f.answer}
                onChange={(v) => onChange(items.map((x, j) => (i === j ? { ...x, answer: v } : x)))}
                rows={3}
              />
            </div>
            <div className="mt-2 flex justify-end">
              <button
                type="button"
                onClick={() => onChange(items.filter((_, j) => j !== i))}
                className="rounded-md p-1.5 text-[var(--dash-text-muted)] hover:bg-[var(--dash-hover)] hover:text-red-400"
              >
                <Trash2 className="h-4 w-4" />
              </button>
            </div>
          </div>
        ))}
        <button
          type="button"
          onClick={() => onChange([...items, { question: '', answer: '' }])}
          className="inline-flex items-center gap-1.5 rounded-md border border-dashed border-[var(--dash-border)] px-3 py-2 text-xs text-[var(--dash-text-muted)] hover:bg-[var(--dash-hover)]"
        >
          <Plus className="h-3 w-3" />
          {t('recruitAdmin.sections.addFaq')}
        </button>
      </div>
    </div>
  );
};

const ValuesEditor: React.FC<{
  items: ValueItem[];
  onChange: (items: ValueItem[]) => void;
}> = ({ items, onChange }) => {
  const { t } = useTranslation('recruitCommon');
  return (
    <div>
      <div className="mb-2 text-xs font-medium uppercase tracking-wider text-[var(--dash-text-muted)]">
        {t('recruitAdmin.sections.values')}
      </div>
      <div className="space-y-2">
        {items.map((v, i) => (
          <div key={i} className="flex gap-2">
            <Input
              value={v.label}
              onChange={(val) =>
                onChange(items.map((x, j) => (i === j ? { ...x, label: val } : x)))
              }
              placeholder={t('recruitAdmin.sections.valueLabelPlaceholder')}
            />
            <Input
              value={v.blurb ?? ''}
              onChange={(val) =>
                onChange(items.map((x, j) => (i === j ? { ...x, blurb: val } : x)))
              }
              placeholder={t('recruitAdmin.sections.valueBlurbPlaceholder')}
            />
            <button
              type="button"
              onClick={() => onChange(items.filter((_, j) => j !== i))}
              className="rounded-md p-2 text-[var(--dash-text-muted)] hover:bg-[var(--dash-hover)] hover:text-red-400"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          </div>
        ))}
        <button
          type="button"
          onClick={() => onChange([...items, { label: '' }])}
          className="inline-flex items-center gap-1.5 rounded-md border border-dashed border-[var(--dash-border)] px-3 py-2 text-xs text-[var(--dash-text-muted)] hover:bg-[var(--dash-hover)]"
        >
          <Plus className="h-3 w-3" />
          {t('recruitAdmin.sections.addValue')}
        </button>
      </div>
    </div>
  );
};

// ── small building-block components ──────────────────────────────────

const Field: React.FC<{
  label: string;
  hint?: string;
  children: React.ReactNode;
}> = ({ label, hint, children }) => (
  <label className="block">
    <div className="mb-1.5 flex items-baseline justify-between">
      <span className="text-xs font-medium uppercase tracking-wider text-[var(--dash-text-muted)]">
        {label}
      </span>
      {hint && <span className="text-[10px] text-[var(--dash-text-dim)]">{hint}</span>}
    </div>
    {children}
  </label>
);

const Input: React.FC<{
  value: string;
  onChange: (v: string) => void;
  type?: string;
  placeholder?: string;
}> = ({ value, onChange, type = 'text', placeholder }) => (
  <input
    type={type}
    value={value}
    onChange={(e) => onChange(e.target.value)}
    placeholder={placeholder}
    className="w-full rounded-md border bg-[var(--dash-hover)] px-3 py-2 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--dash-text)]"
    style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
  />
);

const Textarea: React.FC<{
  value: string;
  onChange: (v: string) => void;
  rows?: number;
}> = ({ value, onChange, rows = 4 }) => (
  <textarea
    value={value}
    onChange={(e) => onChange(e.target.value)}
    rows={rows}
    className="w-full rounded-md border bg-[var(--dash-hover)] px-3 py-2 text-sm leading-relaxed focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--dash-text)]"
    style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
  />
);

export default SectionsAdmin;
