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

/**
 * SectionsAdmin - the editor for the content blocks that sit between
 * the hero and the jobs list on a business's careers page. Laid out
 * as a draggable list of rows; each row expands inline to its editor
 * so the admin keeps context of where it lives in the page.
 *
 * Drag-and-drop uses the native HTML5 drag API (no library) - good
 * enough for a dozen sections, no bundle weight.
 */

// KIND_META is keyed by a narrower set of known kinds than SectionKind
// (which accepts any string for forward-compat). Indexing with a
// SectionKind value works at runtime but TypeScript needs the lookup
// to widen to a plain-string record.
type KnownKind = 'perks' | 'culture' | 'testimonial' | 'video' | 'faq' | 'values' | 'cta';

const KIND_META: Record<KnownKind, { icon: React.ReactNode; label: string; blurb: string }> = {
  perks: {
    icon: <Sparkles className="h-4 w-4" />,
    label: 'Perks',
    blurb: 'Icon grid - healthcare, remote, learning budget, etc.',
  },
  culture: {
    icon: <Users className="h-4 w-4" />,
    label: 'Culture',
    blurb: 'Photo + paragraph. How the team actually works.',
  },
  testimonial: {
    icon: <Quote className="h-4 w-4" />,
    label: 'Testimonial',
    blurb: 'Quote from a current employee with attribution.',
  },
  video: {
    icon: <PlayCircle className="h-4 w-4" />,
    label: 'Video',
    blurb: 'YouTube / Vimeo / Loom embed URL.',
  },
  faq: {
    icon: <HelpCircle className="h-4 w-4" />,
    label: 'FAQ',
    blurb: 'Expandable Q/A list.',
  },
  values: {
    icon: <Star className="h-4 w-4" />,
    label: 'Values',
    blurb: 'Label + one-liner list. What the company stands for.',
  },
  cta: {
    icon: <MousePointerClick className="h-4 w-4" />,
    label: 'Call to action',
    blurb: 'Heading + button pointing at an external URL.',
  },
};

const ORDERED_KINDS: KnownKind[] = [
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
  const { t } = useTranslation();
  const { data, isLoading } = useAdminSections();
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
      { onError: () => toast.error("Couldn't save order") }
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

  return (
    <>
      <HelmetContainer title="Page sections · rtylr" />
      <div className="px-10 py-8">
        <div className="flex items-baseline justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Page sections</h1>
            <p className="mt-1 text-sm text-white/60">
              Content blocks between the hero and the jobs list. Drag to reorder.
            </p>
          </div>
          <button
            type="button"
            onClick={() => setAdding(true)}
            className="inline-flex items-center gap-2 rounded-md bg-white px-4 py-2 text-sm font-semibold text-black hover:bg-white/90"
          >
            <Plus className="h-4 w-4" />
            Add section
          </button>
        </div>

        {adding && (
          <AddSectionPicker
            onCancel={() => setAdding(false)}
            onPick={(kind) => {
              const body: SectionBody = { kind, enabled: true };
              create.mutate(body, {
                onSuccess: (res) => {
                  toast.success('Section added');
                  setAdding(false);
                  if (res?.section) setExpandedId(res.section.id);
                },
                onError: () => toast.error("Couldn't add"),
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
            <div className="p-10 text-center text-sm text-white/60">Loading…</div>
          ) : localOrder.length === 0 ? (
            <EmptyState
              icon={LayoutPanelTop}
              title={t('common.emptyStates.sectionsNone')}
              action={{ label: t('common.emptyStates.addSection'), onClick: () => setAdding(true) }}
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
                  <div className="flex items-center gap-3 px-4 py-3 hover:bg-white/5">
                    <GripVertical
                      className="h-4 w-4 cursor-grab text-white/30 active:cursor-grabbing"
                      aria-hidden
                    />
                    <span className="text-white/60">{lookupMeta(s.kind)?.icon}</span>
                    <div className="min-w-0 flex-1">
                      <div className="truncate text-sm font-medium">
                        {s.title || lookupMeta(s.kind)?.label || s.kind}
                      </div>
                      <div className="truncate text-xs text-white/50">
                        {lookupMeta(s.kind)?.label ?? s.kind}
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
                          { onError: () => toast.error("Couldn't toggle") }
                        )
                      }
                      title={s.enabled ? 'Hide from public page' : 'Show on public page'}
                      className="rounded-md p-1.5 text-white/60 hover:bg-white/10 hover:text-white"
                    >
                      {s.enabled ? <Eye className="h-4 w-4" /> : <EyeOff className="h-4 w-4" />}
                    </button>
                    <button
                      type="button"
                      onClick={() => setExpandedId(expandedId === s.id ? null : s.id)}
                      className="rounded-md p-1.5 text-white/60 hover:bg-white/10 hover:text-white"
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
                        toast.warning('Delete this section?', {
                          action: {
                            label: 'Delete',
                            onClick: () =>
                              del.mutate(s.id, {
                                onSuccess: () => toast.success('Section deleted'),
                                onError: () => toast.error("Couldn't delete"),
                              }),
                          },
                        });
                      }}
                      className="rounded-md p-1.5 text-white/60 hover:bg-white/10 hover:text-red-400"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>

                  {expandedId === s.id && (
                    <div
                      className="border-t bg-black/40 px-4 py-5"
                      style={{ borderColor: 'var(--recruit-admin-border)' }}
                    >
                      <SectionEditor
                        section={s}
                        onSave={(body) =>
                          update.mutate(body, {
                            onSuccess: () => toast.success('Section saved'),
                            onError: () => toast.error("Couldn't save"),
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
}> = ({ onPick, onCancel }) => (
  <div
    className="mt-6 rounded-lg border p-5"
    style={{
      borderColor: 'var(--recruit-admin-border)',
      background: 'var(--recruit-admin-surface)',
    }}
  >
    <div className="flex items-center justify-between">
      <div className="text-sm font-medium">Pick a section kind</div>
      <button
        type="button"
        onClick={onCancel}
        className="rounded-md px-2 py-1 text-xs text-white/60 hover:bg-white/5"
      >
        Cancel
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
            <span className="mb-2 text-white/70">{meta.icon}</span>
            <span className="text-sm font-medium">{meta.label}</span>
            <span className="mt-0.5 text-[10px] text-white/50">{meta.blurb}</span>
          </button>
        );
      })}
    </div>
  </div>
);

// ── SectionEditor - per-kind forms ───────────────────────────────────

const SectionEditor: React.FC<{
  section: Section;
  onSave: (body: SectionBody) => void;
}> = ({ section, onSave }) => {
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
      toast.error(err instanceof Error ? err.message : 'Upload failed');
    } finally {
      setUploadingMedia(false);
      if (mediaInputRef.current) mediaInputRef.current.value = '';
    }
  }

  return (
    <div className="space-y-4">
      <Field label="Section title">
        <Input value={title} onChange={setTitle} placeholder="e.g. Why you'll love it here" />
      </Field>

      {(section.kind === 'culture' || section.kind === 'cta' || !lookupMeta(section.kind)) && (
        <Field label="Body">
          <Textarea value={body} onChange={setBody} rows={4} />
        </Field>
      )}

      {section.kind === 'testimonial' && (
        <>
          <Field label="Quote">
            <Textarea value={body} onChange={setBody} rows={3} />
          </Field>
          <Field label="Attribution" hint="Name, role">
            <Input
              value={mediaUrl}
              onChange={setMediaUrl}
              placeholder="Jamie Doe, Senior Engineer"
            />
          </Field>
        </>
      )}

      {section.kind === 'culture' && (
        <Field label="Photo URL" hint="Optional - square or landscape">
          <div className="flex gap-2">
            <Input value={mediaUrl} onChange={setMediaUrl} type="url" />
            <button
              type="button"
              onClick={() => mediaInputRef.current?.click()}
              disabled={uploadingMedia}
              className="inline-flex items-center gap-2 rounded-md border border-white/10 px-3 text-sm font-semibold text-white/80 hover:bg-white/5 disabled:opacity-60"
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
        <Field label="Video URL" hint="YouTube / Vimeo / Loom or uploaded video">
          <div className="flex gap-2">
            <Input value={mediaUrl} onChange={setMediaUrl} type="url" />
            <button
              type="button"
              onClick={() => mediaInputRef.current?.click()}
              disabled={uploadingMedia}
              className="inline-flex items-center gap-2 rounded-md border border-white/10 px-3 text-sm font-semibold text-white/80 hover:bg-white/5 disabled:opacity-60"
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
        <Field label="Button URL">
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
          className="rounded-md bg-white px-4 py-2 text-sm font-semibold text-black hover:bg-white/90"
        >
          Save section
        </button>
      </div>
    </div>
  );
};

// ── item editors (perks, faq, values) ────────────────────────────────

const PerksEditor: React.FC<{
  items: PerkItem[];
  onChange: (items: PerkItem[]) => void;
}> = ({ items, onChange }) => (
  <div>
    <div className="mb-2 text-xs font-medium uppercase tracking-wider text-white/70">Perks</div>
    <div className="space-y-2">
      {items.map((p, i) => (
        <div key={i} className="flex gap-2">
          <Input
            value={p.icon ?? ''}
            onChange={(v) => onChange(items.map((x, j) => (i === j ? { ...x, icon: v } : x)))}
            placeholder="icon (lucide name)"
          />
          <Input
            value={p.label}
            onChange={(v) => onChange(items.map((x, j) => (i === j ? { ...x, label: v } : x)))}
            placeholder="label"
          />
          <Input
            value={p.blurb ?? ''}
            onChange={(v) => onChange(items.map((x, j) => (i === j ? { ...x, blurb: v } : x)))}
            placeholder="blurb"
          />
          <button
            type="button"
            onClick={() => onChange(items.filter((_, j) => j !== i))}
            className="rounded-md p-2 text-white/60 hover:bg-white/10 hover:text-red-400"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      ))}
      <button
        type="button"
        onClick={() => onChange([...items, { label: '', icon: 'Sparkles' }])}
        className="inline-flex items-center gap-1.5 rounded-md border border-dashed border-white/20 px-3 py-2 text-xs text-white/60 hover:bg-white/5"
      >
        <Plus className="h-3 w-3" />
        Add perk
      </button>
    </div>
  </div>
);

const FaqEditor: React.FC<{
  items: FaqItem[];
  onChange: (items: FaqItem[]) => void;
}> = ({ items, onChange }) => (
  <div>
    <div className="mb-2 text-xs font-medium uppercase tracking-wider text-white/70">Q&A</div>
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
            placeholder="Question"
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
              className="rounded-md p-1.5 text-white/60 hover:bg-white/10 hover:text-red-400"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          </div>
        </div>
      ))}
      <button
        type="button"
        onClick={() => onChange([...items, { question: '', answer: '' }])}
        className="inline-flex items-center gap-1.5 rounded-md border border-dashed border-white/20 px-3 py-2 text-xs text-white/60 hover:bg-white/5"
      >
        <Plus className="h-3 w-3" />
        Add Q/A
      </button>
    </div>
  </div>
);

const ValuesEditor: React.FC<{
  items: ValueItem[];
  onChange: (items: ValueItem[]) => void;
}> = ({ items, onChange }) => (
  <div>
    <div className="mb-2 text-xs font-medium uppercase tracking-wider text-white/70">Values</div>
    <div className="space-y-2">
      {items.map((v, i) => (
        <div key={i} className="flex gap-2">
          <Input
            value={v.label}
            onChange={(val) => onChange(items.map((x, j) => (i === j ? { ...x, label: val } : x)))}
            placeholder="label"
          />
          <Input
            value={v.blurb ?? ''}
            onChange={(val) => onChange(items.map((x, j) => (i === j ? { ...x, blurb: val } : x)))}
            placeholder="short description"
          />
          <button
            type="button"
            onClick={() => onChange(items.filter((_, j) => j !== i))}
            className="rounded-md p-2 text-white/60 hover:bg-white/10 hover:text-red-400"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      ))}
      <button
        type="button"
        onClick={() => onChange([...items, { label: '' }])}
        className="inline-flex items-center gap-1.5 rounded-md border border-dashed border-white/20 px-3 py-2 text-xs text-white/60 hover:bg-white/5"
      >
        <Plus className="h-3 w-3" />
        Add value
      </button>
    </div>
  </div>
);

// ── small building-block components ──────────────────────────────────

const Field: React.FC<{
  label: string;
  hint?: string;
  children: React.ReactNode;
}> = ({ label, hint, children }) => (
  <label className="block">
    <div className="mb-1.5 flex items-baseline justify-between">
      <span className="text-xs font-medium uppercase tracking-wider text-white/70">{label}</span>
      {hint && <span className="text-[10px] text-white/40">{hint}</span>}
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
    className="w-full rounded-md border bg-black/40 px-3 py-2 text-sm outline-none focus:border-white/40"
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
    className="w-full rounded-md border bg-black/40 px-3 py-2 text-sm leading-relaxed outline-none focus:border-white/40"
    style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
  />
);

export default SectionsAdmin;
