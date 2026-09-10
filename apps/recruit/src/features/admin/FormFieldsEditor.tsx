import React, { useState } from 'react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import {
  Plus,
  Trash2,
  Type as TypeIcon,
  AlignLeft,
  List,
  CheckSquare,
  Link as LinkIcon,
  ToggleRight,
  HelpCircle,
} from 'lucide-react';
import { EmptyState } from '@voxire/ui';

import {
  useAdminFormFields,
  useCreateFormField,
  useUpdateFormField,
  useDeleteFormField,
  type FormFieldBody,
} from '../../shared/hooks/queries/useAdmin';
import type { FormField, FormFieldKind } from '../../shared/lib/types';

/**
 * FormFieldsEditor - lets the HR admin add custom questions to the
 * apply form for a specific role. Baseline (name/email/resume/cover
 * letter) isn't editable; everything here appears after it on the
 * public apply page.
 *
 * The editor is inline rather than its own route because it's always
 * edited in the context of a role - bouncing to a separate screen
 * would make the connection between question and role unclear.
 */

const KIND_META: Record<
  FormFieldKind,
  { icon: React.ReactNode; label: string; hasOptions: boolean }
> = {
  short_text: {
    icon: <TypeIcon className="h-3.5 w-3.5" />,
    label: 'Short text',
    hasOptions: false,
  },
  long_text: {
    icon: <AlignLeft className="h-3.5 w-3.5" />,
    label: 'Long text',
    hasOptions: false,
  },
  single_select: {
    icon: <List className="h-3.5 w-3.5" />,
    label: 'Dropdown',
    hasOptions: true,
  },
  multi_select: {
    icon: <CheckSquare className="h-3.5 w-3.5" />,
    label: 'Multi-choice',
    hasOptions: true,
  },
  url: { icon: <LinkIcon className="h-3.5 w-3.5" />, label: 'URL', hasOptions: false },
  boolean: { icon: <ToggleRight className="h-3.5 w-3.5" />, label: 'Checkbox', hasOptions: false },
};

const KIND_ORDER: FormFieldKind[] = [
  'short_text',
  'long_text',
  'single_select',
  'multi_select',
  'url',
  'boolean',
];

interface Props {
  careerId: number;
}

const FormFieldsEditor: React.FC<Props> = ({ careerId }) => {
  const { t } = useTranslation();
  const { data, isLoading } = useAdminFormFields(careerId);
  const create = useCreateFormField();
  const update = useUpdateFormField();
  const del = useDeleteFormField();

  const fields = data?.fields ?? [];

  function addDefaultQuestion(): void {
    const kind = KIND_ORDER[0];
    create.mutate(
      { careerId, kind, label: defaultLabelFor(kind), required: false },
      {
        onSuccess: () => toast.success(`${KIND_META[kind].label} added`),
        onError: () => toast.error("Couldn't add"),
      }
    );
  }

  return (
    <div
      className="rounded-lg border p-5"
      style={{
        borderColor: 'var(--recruit-admin-border-strong)',
        background: 'rgba(255,255,255,0.02)',
      }}
    >
      <div className="mb-4 flex items-center justify-between">
        <div>
          <div className="text-sm font-semibold">Custom questions</div>
          <div className="text-xs text-white/50">
            Asked after the default name / email / resume block on the apply form.
          </div>
        </div>
      </div>

      {isLoading ? (
        <div className="py-6 text-center text-sm text-white/50">Loading…</div>
      ) : fields.length === 0 ? (
        <div className="mb-4">
          <EmptyState
            icon={HelpCircle}
            title={t('common.emptyStates.questionsNone')}
            action={{ label: t('common.emptyStates.addQuestion'), onClick: addDefaultQuestion }}
            size="compact"
          />
        </div>
      ) : (
        <ul className="mb-4 space-y-2">
          {fields.map((f) => (
            <FormFieldRow
              key={f.id}
              field={f}
              onSave={(body) =>
                update.mutate(body, {
                  onSuccess: () => toast.success('Question saved'),
                  onError: () => toast.error("Couldn't save"),
                })
              }
              onDelete={() =>
                del.mutate(
                  { id: f.id, careerId },
                  {
                    onSuccess: () => toast.success('Question removed'),
                    onError: () => toast.error("Couldn't delete"),
                  }
                )
              }
            />
          ))}
        </ul>
      )}

      <div className="flex flex-wrap gap-1.5">
        {KIND_ORDER.map((k) => {
          const meta = KIND_META[k];
          return (
            <button
              key={k}
              type="button"
              onClick={() =>
                create.mutate(
                  {
                    careerId,
                    kind: k,
                    label: defaultLabelFor(k),
                    required: false,
                    optionsJson:
                      k === 'single_select' || k === 'multi_select'
                        ? JSON.stringify(['Option 1', 'Option 2'])
                        : undefined,
                  },
                  {
                    onSuccess: () => toast.success(`${meta.label} added`),
                    onError: () => toast.error("Couldn't add"),
                  }
                )
              }
              className="inline-flex items-center gap-1.5 rounded-md border border-white/10 px-2.5 py-1 text-xs hover:bg-white/5"
            >
              {meta.icon}
              <Plus className="h-3 w-3" />
              {meta.label}
            </button>
          );
        })}
      </div>
    </div>
  );
};

function defaultLabelFor(kind: FormFieldKind): string {
  switch (kind) {
    case 'url':
      return 'Portfolio URL';
    case 'long_text':
      return 'Tell us more';
    case 'single_select':
      return 'Pick one';
    case 'multi_select':
      return 'Pick all that apply';
    case 'boolean':
      return 'I agree';
    case 'short_text':
    default:
      return 'New question';
  }
}

// ── per-row editor ────────────────────────────────────────────────────

const FormFieldRow: React.FC<{
  field: FormField;
  onSave: (body: FormFieldBody) => void;
  onDelete: () => void;
}> = ({ field, onSave, onDelete }) => {
  const meta = (KIND_META as Record<string, (typeof KIND_META)[FormFieldKind]>)[field.kind];
  const [label, setLabel] = useState(field.label);
  const [placeholder, setPlaceholder] = useState(field.placeholder);
  const [required, setRequired] = useState(field.required);
  const [options, setOptions] = useState<string[]>(() => parseOptions(field.optionsJson));

  const hasOptions = meta?.hasOptions;

  function commit() {
    onSave({
      id: field.id,
      careerId: field.careerId,
      kind: field.kind,
      label,
      placeholder,
      required,
      optionsJson: hasOptions && options.length > 0 ? JSON.stringify(options) : '',
      position: field.position,
    });
  }

  return (
    <li
      className="rounded-md border p-3"
      style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
    >
      <div className="flex items-center gap-2">
        <span className="inline-flex items-center gap-1 rounded-md bg-white/5 px-2 py-1 text-xs text-white/70">
          {meta?.icon}
          {meta?.label ?? field.kind}
        </span>
        <input
          value={label}
          onChange={(e) => setLabel(e.target.value)}
          onBlur={commit}
          className="flex-1 rounded-md border bg-black/40 px-2.5 py-1.5 text-sm outline-none focus:border-white/40"
          style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
          placeholder="Question label"
        />
        <label className="inline-flex cursor-pointer items-center gap-1 text-xs text-white/70">
          <input
            type="checkbox"
            checked={required}
            onChange={(e) => {
              const next = e.target.checked;
              setRequired(next);
              onSave({
                id: field.id,
                careerId: field.careerId,
                kind: field.kind,
                label,
                placeholder,
                required: next,
                optionsJson: hasOptions && options.length > 0 ? JSON.stringify(options) : '',
                position: field.position,
              });
            }}
            className="h-3.5 w-3.5"
          />
          Required
        </label>
        <button
          type="button"
          onClick={() => {
            toast.warning('Remove this field?', {
              action: { label: 'Remove', onClick: onDelete },
            });
          }}
          className="rounded-md p-1.5 text-white/60 hover:bg-white/10 hover:text-red-400"
        >
          <Trash2 className="h-3.5 w-3.5" />
        </button>
      </div>

      {(field.kind === 'short_text' || field.kind === 'long_text' || field.kind === 'url') && (
        <div className="mt-2">
          <input
            value={placeholder}
            onChange={(e) => setPlaceholder(e.target.value)}
            onBlur={commit}
            className="w-full rounded-md border bg-black/40 px-2.5 py-1.5 text-xs outline-none focus:border-white/40"
            style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
            placeholder="Placeholder (optional)"
          />
        </div>
      )}

      {hasOptions && (
        <div className="mt-2 space-y-1.5">
          {options.map((opt, i) => (
            <div key={i} className="flex gap-2">
              <input
                value={opt}
                onChange={(e) => {
                  const next = options.map((x, j) => (i === j ? e.target.value : x));
                  setOptions(next);
                }}
                onBlur={commit}
                className="flex-1 rounded-md border bg-black/40 px-2.5 py-1.5 text-xs outline-none focus:border-white/40"
                style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
                placeholder={`Option ${i + 1}`}
              />
              <button
                type="button"
                onClick={() => {
                  const next = options.filter((_, j) => j !== i);
                  setOptions(next);
                  onSave({
                    id: field.id,
                    careerId: field.careerId,
                    kind: field.kind,
                    label,
                    placeholder,
                    required,
                    optionsJson: next.length > 0 ? JSON.stringify(next) : '',
                    position: field.position,
                  });
                }}
                className="rounded-md p-1.5 text-white/60 hover:bg-white/10 hover:text-red-400"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </button>
            </div>
          ))}
          <button
            type="button"
            onClick={() => {
              const next = [...options, ''];
              setOptions(next);
            }}
            className="inline-flex items-center gap-1 rounded-md border border-dashed border-white/20 px-2.5 py-1 text-xs text-white/60 hover:bg-white/5"
          >
            <Plus className="h-3 w-3" />
            Add option
          </button>
        </div>
      )}
    </li>
  );
};

function parseOptions(json: string): string[] {
  if (!json) return [];
  try {
    const parsed = JSON.parse(json);
    return Array.isArray(parsed) ? parsed.map(String) : [];
  } catch {
    return [];
  }
}

export default FormFieldsEditor;
