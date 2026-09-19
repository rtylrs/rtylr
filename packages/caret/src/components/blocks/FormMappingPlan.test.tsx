import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { ToolCallBlock } from '../../types/message';
import { decodeFormMappingPlan, FormMappingPlan } from './FormMappingPlan';
import ToolCallCard from './ToolCallCard';

const planOutput = {
  summary: 'Planned 3 of 4 fields from 3 extracted values (2 fills, 1 checks, 1 skips, 1 need review). Scorer: s1-forms.',
  scorer: 's1-forms',
  field_count: 4,
  fill_count: 2,
  check_count: 1,
  skip_count: 1,
  needs_review_count: 1,
  extracted_entity_count: 3,
  apply_order: 'fills, then checks, then any approved submit click; skips are not applied',
  fields: [
    { id: 'phone', label: 'Phone number', action: 'fill', value: '(503) 555-0142', option_index: 1, probability: 0.92 },
    { id: 'insurance', label: 'Has insurance', action: 'check', option_index: 3, probability: 0.8 },
    { id: 'notes', label: 'Notes', action: 'skip', option_index: -1, probability: 0.4, needs_review: true },
    { id: 'name', label: 'Legal name', action: 'skip', option_index: -1, probability: 0, already_set: true },
  ],
  submit: { label: 'Submit', action: 'click', probability: 0.93, requires_approval: true },
};

describe('FormMappingPlan', () => {
  it('decodes the tool output defensively', () => {
    expect(decodeFormMappingPlan(undefined)).toBeNull();
    expect(decodeFormMappingPlan('nope')).toBeNull();
    expect(decodeFormMappingPlan([])).toBeNull();
    expect(decodeFormMappingPlan({})).toBeNull();
    expect(decodeFormMappingPlan({ fields: 'nope' })).toBeNull();
    expect(decodeFormMappingPlan({ fields: [{ label: 'x' }] })?.fields).toEqual([]);

    const plan = decodeFormMappingPlan(planOutput);
    expect(plan?.fields).toHaveLength(4);
    expect(plan?.fields[0]).toMatchObject({
      label: 'Phone number',
      action: 'fill',
      value: '(503) 555-0142',
      probability: 0.92,
    });
    expect(plan?.submit).toMatchObject({ label: 'Submit', requires_approval: true });
  });

  it('renders rows, review flags, and the submit block', () => {
    render(<FormMappingPlan output={planOutput} />);

    expect(screen.getByText('Phone number')).not.toBeNull();
    expect(screen.getByText(/\(503\) 555-0142/)).not.toBeNull();
    expect(screen.getByText(/92%/)).not.toBeNull();
    expect(screen.getByText('Has insurance')).not.toBeNull();
    expect(screen.getByText('needs review')).not.toBeNull();
    expect(screen.getByText('already set')).not.toBeNull();
    expect(screen.getByText('Submit')).not.toBeNull();
    expect(screen.getByText(/needs approval/)).not.toBeNull();
    expect(screen.getByText('model')).not.toBeNull();
  });

  it('renders nothing for missing or malformed output', () => {
    for (const output of [undefined, 'nope', [], {}, { fields: 'nope' }]) {
      const { container } = render(<FormMappingPlan output={output} />);
      expect(container.firstChild).toBeNull();
    }
  });
});

describe('ToolCallCard with documents.map_form_fields', () => {
  it('shows the structured plan inside the expanded card', () => {
    const call: ToolCallBlock = {
      id: 'tool-plan',
      type: 'tool_call',
      name: 'documents.map_form_fields',
      input: { document: 'Phone: (503) 555-0142', fields: [{ label: 'Phone' }] },
      status: 'success',
      output: planOutput,
      summary: 'Planned 3 of 4 fields.',
    };

    render(<ToolCallCard call={call} />);
    fireEvent.click(screen.getByRole('button', { name: /Completed: map a document onto the form/i }));

    expect(screen.getByText('Phone number')).not.toBeNull();
    expect(screen.getByText(/\(503\) 555-0142/)).not.toBeNull();
    expect(screen.getByText('Planned 3 of 4 fields.')).not.toBeNull();
  });

  it('does not render the plan for other tools', () => {
    const call: ToolCallBlock = {
      id: 'tool-other',
      type: 'tool_call',
      name: 'insights.get_product_kpis',
      input: {},
      status: 'success',
      output: planOutput,
      summary: 'Checked product data.',
    };

    render(<ToolCallCard call={call} />);
    fireEvent.click(screen.getByRole('button', { name: /Completed: check business data/i }));

    expect(screen.queryByText('Phone number')).toBeNull();
    expect(screen.queryByText(/needs approval/)).toBeNull();
  });
});
