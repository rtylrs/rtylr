import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Plus, X } from 'lucide-react';
import { Switch } from '@voxire/ui';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../../../../../shared/components/ui/dialog';
import { Button } from '../../../../../shared/components/ui/button';
import type {
  ApiField,
  ApiFieldInput,
  ApiFieldType,
  ApiTaskFieldValue,
} from '../../../shared/services/flowApi';
import CustomFieldInput, {
  type CustomFieldPerson,
} from '../../../shared/components/CustomFieldInput';
import FlowSelect from '../../../shared/components/FlowSelect';
import { FIELD_TYPES, hasOptions, isMultiField } from '../../../shared/lib/customFields';
import { STATUS_COLORS } from '../statusUtils';

interface DraftOption {
  /** The server's id, or a local `new:<n>` key until it has one. */
  key: string;
  id?: string;
  label: string;
  color: string;
}

export interface FieldEditorSubmit extends ApiFieldInput {
  type: ApiFieldType;
}

interface FieldEditorDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** The field being edited; absent to create one. */
  field?: ApiField;
  people: readonly CustomFieldPerson[];
  pending: boolean;
  onSubmit: (data: FieldEditorSubmit) => void;
}

let nextKey = 0;
const newOption = (label = ''): DraftOption => ({
  key: `new:${nextKey++}`,
  label,
  color: STATUS_COLORS[nextKey % STATUS_COLORS.length] ?? '#6366f1',
});

/**
 * Create or edit one custom field: name, type (fixed once created), choices
 * for a select, the required flag and a default. Choices keep their id when
 * renamed, so renaming never loses a task's value.
 */
const FieldEditorDialog: React.FC<FieldEditorDialogProps> = ({
  open,
  onOpenChange,
  field,
  people,
  pending,
  onSubmit,
}) => {
  const { t, i18n } = useTranslation();
  const [name, setName] = useState('');
  const [type, setType] = useState<ApiFieldType>('text');
  const [options, setOptions] = useState<DraftOption[]>([]);
  const [required, setRequired] = useState(false);
  const [defaultValue, setDefaultValue] = useState<ApiTaskFieldValue>({});

  useEffect(() => {
    if (!open) return;
    setName(field?.name ?? '');
    setType(field?.type ?? 'text');
    setOptions(
      field?.options.map((o) => ({ key: o.id, id: o.id, label: o.label, color: o.color })) ?? []
    );
    setRequired(field?.required ?? false);
    setDefaultValue(
      field
        ? isMultiField(field.type)
          ? { values: field.defaultValues }
          : { value: field.defaultValue }
        : {}
    );
  }, [open, field]);

  // The default is picked with the same input a task uses, so it needs a field
  // shaped like the draft. New choices stand in under their local key.
  const draftField = useMemo<ApiField>(
    () => ({
      id: field?.id ?? 'draft',
      projectId: field?.projectId ?? '',
      name: t('flow.fields.defaultValue'),
      type,
      options: options
        .filter((o) => o.label.trim() !== '')
        .map((o) => ({ id: o.key, label: o.label.trim(), color: o.color })),
      required: false,
      defaultValue: '',
      defaultValues: [],
      sortOrder: 0,
      valueCount: 0,
    }),
    [field, type, options, t]
  );

  const labels = options.map((o) => o.label.trim().toLowerCase()).filter(Boolean);
  const duplicateLabel = new Set(labels).size !== labels.length;
  const needsOptions = hasOptions(type) && labels.length === 0;
  const canSave = name.trim() !== '' && !duplicateLabel && !needsOptions && !pending;

  const submit = (): void => {
    if (!canSave) return;
    const kept = options.filter((o) => o.label.trim() !== '');
    // A default naming a new choice is sent as its label: the server resolves
    // it once the choice has an id.
    const resolve = (key: string): string => {
      const option = kept.find((o) => o.key === key);
      return option ? (option.id ?? option.label.trim()) : key;
    };
    const live = (key: string): boolean => !hasOptions(type) || kept.some((o) => o.key === key);
    onSubmit({
      type,
      name: name.trim(),
      options: hasOptions(type)
        ? kept.map((o) => ({ id: o.id, label: o.label.trim(), color: o.color }))
        : [],
      required,
      defaultValue:
        !isMultiField(type) && defaultValue.value && live(defaultValue.value)
          ? resolve(defaultValue.value)
          : '',
      defaultValues: isMultiField(type)
        ? (defaultValue.values ?? []).filter(live).map(resolve)
        : [],
    });
  };

  const updateOption = (key: string, patch: Partial<DraftOption>): void =>
    setOptions((current) => current.map((o) => (o.key === key ? { ...o, ...patch } : o)));

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="portal-flow-dialog flex flex-col sm:max-w-lg"
        dir={i18n.dir()}
        closeLabel={t('flowCommon.close')}
      >
        <DialogHeader className="shrink-0">
          <DialogTitle>
            {field ? t('flow.fields.editTitle') : t('flow.fields.addTitle')}
          </DialogTitle>
          <DialogDescription>{t('flow.fields.editorHint')}</DialogDescription>
        </DialogHeader>

        <form
          className="min-h-0 flex-1 space-y-4 overflow-y-auto px-6 pb-2"
          onSubmit={(e) => {
            e.preventDefault();
            submit();
          }}
        >
          <label className="block">
            <span
              className="mb-1 block text-flow-meta font-medium"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flowCommon.name')}
            </span>
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={64}
              autoFocus
              aria-label={t('flow.fields.fieldName')}
              className="w-full rounded-md px-3 py-2 text-flow-meta"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border-control)',
                color: 'var(--flow-text)',
              }}
            />
          </label>

          <div>
            <span
              className="mb-1 block text-flow-meta font-medium"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.fields.type')}
            </span>
            <FlowSelect
              label={t('flow.fields.type')}
              value={type}
              disabled={Boolean(field)}
              options={FIELD_TYPES.map((ft) => ({
                value: ft,
                label: t(`flow.fields.types.${ft}`),
              }))}
              onChange={(next) => {
                setType(next as ApiFieldType);
                setDefaultValue({});
                if (hasOptions(next as ApiFieldType) && options.length === 0) {
                  setOptions([newOption()]);
                }
              }}
            />
            {field && (
              <p className="mt-1 text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
                {t('flow.fields.typeFixed')}
              </p>
            )}
          </div>

          {hasOptions(type) && (
            <fieldset>
              <legend
                className="mb-1 text-flow-meta font-medium"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.fields.options')}
              </legend>
              <ul className="space-y-1.5">
                {options.map((o, i) => (
                  <li key={o.key} className="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() => {
                        const at = STATUS_COLORS.indexOf(o.color);
                        updateOption(o.key, {
                          color: STATUS_COLORS[(at + 1) % STATUS_COLORS.length] ?? o.color,
                        });
                      }}
                      aria-label={t('flow.fields.changeColor', { label: o.label || i + 1 })}
                      className="flow-focus h-5 w-5 shrink-0 rounded-full"
                      style={{ background: o.color }}
                    />
                    <input
                      value={o.label}
                      maxLength={64}
                      onChange={(e) => updateOption(o.key, { label: e.target.value })}
                      aria-label={t('flow.fields.optionLabel', { index: i + 1 })}
                      className="min-w-0 flex-1 rounded-md px-2.5 py-1.5 text-flow-meta"
                      style={{
                        background: 'var(--flow-input-bg)',
                        border: '1px solid var(--flow-border-control)',
                        color: 'var(--flow-text)',
                      }}
                    />
                    <button
                      type="button"
                      onClick={() =>
                        setOptions((current) => current.filter((x) => x.key !== o.key))
                      }
                      aria-label={t('flow.fields.removeOption', { label: o.label || i + 1 })}
                      className="flow-focus flex h-8 w-8 shrink-0 items-center justify-center rounded-sm"
                      style={{ color: 'var(--flow-text-muted)' }}
                    >
                      <X size={14} aria-hidden="true" />
                    </button>
                  </li>
                ))}
              </ul>
              <button
                type="button"
                onClick={() => setOptions((current) => [...current, newOption()])}
                className="flow-focus mt-2 inline-flex items-center gap-1 rounded-sm px-2 py-1 text-flow-meta"
                style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
              >
                <Plus size={12} aria-hidden="true" /> {t('flow.fields.addOption')}
              </button>
              {field && field.options.some((o) => !options.some((d) => d.id === o.id)) && (
                <p
                  role="status"
                  className="mt-2 text-flow-meta"
                  style={{ color: 'var(--flow-warn-fg)' }}
                >
                  {t('flow.fields.removedOptionsWarning')}
                </p>
              )}
              {duplicateLabel && (
                <p
                  role="alert"
                  className="mt-2 text-flow-meta"
                  style={{ color: 'var(--flow-danger-fg)' }}
                >
                  {t('flow.fields.duplicateOption')}
                </p>
              )}
            </fieldset>
          )}

          <label className="flex items-center justify-between gap-3">
            <span className="text-flow-meta" style={{ color: 'var(--flow-text)' }}>
              {t('flow.fields.requiredLabel')}
              <span className="block text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
                {t('flow.fields.requiredHint')}
              </span>
            </span>
            <Switch
              checked={required}
              onCheckedChange={setRequired}
              aria-label={t('flow.fields.requiredLabel')}
            />
          </label>

          <div>
            <span
              className="mb-1 block text-flow-meta font-medium"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.fields.defaultValue')}
            </span>
            <CustomFieldInput
              key={type}
              field={draftField}
              value={defaultValue}
              people={people}
              onChange={setDefaultValue}
            />
          </div>
        </form>

        <DialogFooter className="shrink-0">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('flowCommon.cancel')}
          </Button>
          <Button onClick={submit} disabled={!canSave}>
            {pending
              ? t('flowCommon.saving')
              : field
                ? t('flowCommon.saveChanges')
                : t('flow.fields.addField')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default FieldEditorDialog;
