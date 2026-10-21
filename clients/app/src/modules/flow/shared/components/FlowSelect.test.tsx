import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import FlowSelect from './FlowSelect';
import { chooseFlowOption, flowSelectTrigger } from '../../../../../test/flowSelect';

const options = [
  { value: 'urgent', label: 'Urgent' },
  { value: 'high', label: 'High' },
  { value: 'low', label: 'Low' },
];

const renderSelect = (props: Partial<React.ComponentProps<typeof FlowSelect>> = {}) => {
  const onChange = vi.fn();
  render(
    <FlowSelect label="Priority" value="high" options={options} onChange={onChange} {...props} />
  );
  return onChange;
};

describe('FlowSelect', () => {
  it('names the control after its label and exposes the current value', () => {
    renderSelect();
    expect(flowSelectTrigger('Priority')).toHaveAccessibleName('Priority: High');
  });

  it('opens on pointer down and reports the chosen value', async () => {
    const onChange = renderSelect();
    await chooseFlowOption('Priority', 'Urgent');
    expect(onChange).toHaveBeenCalledExactlyOnceWith('urgent');
  });

  it('marks the current option and does not report re-selecting it', async () => {
    const onChange = renderSelect();
    fireEvent.pointerDown(flowSelectTrigger('Priority'), { button: 0, ctrlKey: false });
    const current = await screen.findByRole('menuitem', { name: 'High' });
    expect(current).toHaveAttribute('aria-current', 'true');
    fireEvent.click(current);
    expect(onChange).not.toHaveBeenCalled();
  });

  it('does not open when disabled', () => {
    renderSelect({ disabled: true });
    const trigger = flowSelectTrigger('Priority');
    expect(trigger).toBeDisabled();
    fireEvent.pointerDown(trigger, { button: 0, ctrlKey: false });
    expect(screen.queryByRole('menu')).not.toBeInTheDocument();
  });

  it('falls back to the placeholder when no option matches', () => {
    renderSelect({ value: '', placeholder: 'Pick one' });
    expect(flowSelectTrigger('Priority')).toHaveAccessibleName('Priority: Pick one');
  });
});
