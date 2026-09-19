import React from 'react';
import {
  AlertCircle,
  CircleCheck,
  ClipboardCheck,
  ListChecks,
  MousePointerClick,
  PenLine,
  SkipForward,
  TriangleAlert,
} from 'lucide-react';
import { caretTokens, caretRadius } from '../../ui/tokens';

export interface FormMappingField {
  id?: string;
  label: string;
  action: string;
  value?: string;
  probability?: number;
  needs_review?: boolean;
  already_set?: boolean;
}

export interface FormMappingSubmit {
  label: string;
  action: string;
  probability?: number;
  requires_approval?: boolean;
}

export interface FormMappingPlanData {
  summary?: string;
  scorer?: string;
  field_count?: number;
  fill_count?: number;
  check_count?: number;
  skip_count?: number;
  needs_review_count?: number;
  extracted_entity_count?: number;
  fields: FormMappingField[];
  submit?: FormMappingSubmit;
  notice?: string;
  apply_order?: string;
}

export interface FormMappingPlanProps {
  output: unknown;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function asText(value: unknown): string {
  return typeof value === 'string' ? value : '';
}

function asNumber(value: unknown): number | undefined {
  return typeof value === 'number' && Number.isFinite(value) ? value : undefined;
}

function asFlag(value: unknown): boolean {
  return value === true;
}

function decodeField(entry: unknown): FormMappingField | null {
  if (!isRecord(entry)) return null;
  const label = asText(entry.label);
  const action = asText(entry.action);
  if (!label || !action) return null;
  return {
    id: asText(entry.id) || undefined,
    label,
    action,
    value: asText(entry.value) || undefined,
    probability: asNumber(entry.probability),
    needs_review: asFlag(entry.needs_review),
    already_set: asFlag(entry.already_set),
  };
}

function decodeSubmit(entry: unknown): FormMappingSubmit | null {
  if (!isRecord(entry)) return null;
  const label = asText(entry.label);
  const action = asText(entry.action);
  if (!label || !action) return null;
  return {
    label,
    action,
    probability: asNumber(entry.probability),
    requires_approval: asFlag(entry.requires_approval),
  };
}

export function decodeFormMappingPlan(output: unknown): FormMappingPlanData | null {
  if (!isRecord(output)) return null;
  if (!Array.isArray(output.fields)) return null;
  const fields: FormMappingField[] = [];
  for (const entry of output.fields) {
    const field = decodeField(entry);
    if (field) fields.push(field);
  }
  return {
    summary: asText(output.summary) || undefined,
    scorer: asText(output.scorer) || undefined,
    field_count: asNumber(output.field_count),
    fill_count: asNumber(output.fill_count),
    check_count: asNumber(output.check_count),
    skip_count: asNumber(output.skip_count),
    needs_review_count: asNumber(output.needs_review_count),
    extracted_entity_count: asNumber(output.extracted_entity_count),
    fields,
    submit: decodeSubmit(output.submit) ?? undefined,
    notice: asText(output.notice) || undefined,
    apply_order: asText(output.apply_order) || undefined,
  };
}

function actionIcon(action: string): React.ReactNode {
  const style = { flex: '0 0 auto', marginTop: 1 } as const;
  switch (action) {
    case 'fill':
      return <PenLine size={12} style={style} />;
    case 'check':
      return <ClipboardCheck size={12} style={style} />;
    case 'click':
      return <MousePointerClick size={12} style={style} />;
    case 'skip':
      return <SkipForward size={12} style={style} />;
    default:
      return <AlertCircle size={12} style={style} />;
  }
}

function formatProbability(probability: number | undefined): string {
  if (probability === undefined) return '';
  return `${Math.round(probability * 100)}%`;
}

function truncate(value: string, max: number): string {
  if (value.length <= max) return value;
  return `${value.slice(0, max - 1)}…`;
}

export const FormMappingPlan: React.FC<FormMappingPlanProps> = ({ output }) => {
  const plan = decodeFormMappingPlan(output);
  if (!plan) return null;

  const parts: string[] = [];
  if (plan.fill_count !== undefined) parts.push(`${plan.fill_count} fills`);
  if (plan.check_count !== undefined) parts.push(`${plan.check_count} checks`);
  if (plan.skip_count !== undefined) parts.push(`${plan.skip_count} skips`);
  if (plan.needs_review_count !== undefined && plan.needs_review_count > 0) {
    parts.push(`${plan.needs_review_count} need review`);
  }

  return (
    <div style={{ marginBottom: 8 }}>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 6,
          color: caretTokens.textDim,
          fontSize: 11,
          marginBottom: 4,
        }}
      >
        <ListChecks size={12} style={{ flex: '0 0 auto' }} />
        <span>form plan{parts.length > 0 ? ` · ${parts.join(' · ')}` : ''}</span>
        {plan.scorer && (
          <span
            style={{
              marginLeft: 'auto',
              padding: '1px 6px',
              borderRadius: caretRadius.sm,
              border: `1px solid ${caretTokens.border}`,
              color: caretTokens.textDim,
              fontSize: 10,
              whiteSpace: 'nowrap',
            }}
          >
            {plan.scorer === 's1-forms' ? 'model' : plan.scorer}
          </span>
        )}
      </div>

      <div
        style={{
          border: `1px solid ${caretTokens.border}`,
          borderRadius: caretRadius.sm,
          background: caretTokens.surfaceHigh,
          overflow: 'hidden',
          fontFamily: 'inherit',
        }}
      >
        {plan.fields.map((field, index) => (
          <div
            key={field.id ?? `${field.label}-${index}`}
            style={{
              display: 'flex',
              alignItems: 'flex-start',
              gap: 7,
              padding: '6px 9px',
              borderTop: index === 0 ? 'none' : `1px solid ${caretTokens.border}`,
              color: caretTokens.textMuted,
              fontSize: 12,
              lineHeight: 1.45,
            }}
          >
            {actionIcon(field.action)}
            <div style={{ minWidth: 0, flex: '1 1 auto' }}>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'baseline',
                  gap: 6,
                  minWidth: 0,
                }}
              >
                <span
                  style={{
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                    color: caretTokens.text,
                  }}
                >
                  {field.label}
                </span>
                <span style={{ color: caretTokens.textDim, fontSize: 11, flex: '0 0 auto' }}>
                  {field.action}
                  {formatProbability(field.probability)
                    ? ` ${formatProbability(field.probability)}`
                    : ''}
                </span>
              </div>
              {field.action === 'fill' && field.value && (
                <div
                  style={{
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                    color: caretTokens.textMuted,
                    fontSize: 11,
                  }}
                >
                  → {truncate(field.value, 80)}
                </div>
              )}
              <div style={{ display: 'flex', gap: 6, marginTop: 2 }}>
                {field.needs_review && (
                  <span
                    style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: 3,
                      color: caretTokens.warn,
                      fontSize: 10,
                    }}
                  >
                    <TriangleAlert size={10} /> needs review
                  </span>
                )}
                {field.already_set && (
                  <span
                    style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: 3,
                      color: caretTokens.textDim,
                      fontSize: 10,
                    }}
                  >
                    <CircleCheck size={10} /> already set
                  </span>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      {plan.submit && (
        <div
          style={{
            display: 'flex',
            alignItems: 'flex-start',
            gap: 6,
            marginTop: 6,
            padding: '6px 9px',
            border: `1px solid ${caretTokens.accentBorder}`,
            borderRadius: caretRadius.sm,
            background: caretTokens.accentSoft,
            color: caretTokens.text,
            fontSize: 12,
          }}
        >
          <MousePointerClick size={12} style={{ flex: '0 0 auto', marginTop: 2 }} />
          <span>
            {plan.submit.label}
            {plan.submit.requires_approval && (
              <span style={{ color: caretTokens.warn }}> · needs approval</span>
            )}
          </span>
        </div>
      )}

      {plan.notice && (
        <div style={{ marginTop: 6, color: caretTokens.warn, fontSize: 11, lineHeight: 1.45 }}>
          {plan.notice}
        </div>
      )}
    </div>
  );
};

export default FormMappingPlan;
