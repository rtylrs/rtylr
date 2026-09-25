import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { useCaretSettingsStore } from '../stores/settingsStore';
import SettingsPanel from './SettingsPanel';

describe('SettingsPanel', () => {
  beforeEach(() => {
    useCaretSettingsStore.setState({
      mode: 'ask',
      voiceEnabled: true,
      autoApproveReadOnly: true,
      showThinking: true,
    });
  });

  it('updates the default mode from the native select', () => {
    render(<SettingsPanel />);

    fireEvent.change(screen.getByRole('combobox', { name: 'Response style' }), {
      target: { value: 'analyst' },
    });

    expect(useCaretSettingsStore.getState().mode).toBe('analyst');
    expect(screen.getByText('Reports, tables, and KPI breakdowns. Read-only.')).toBeTruthy();
  });

  it('uses full-row accessible switches for behavior settings', () => {
    render(<SettingsPanel />);

    const voiceSwitch = screen.getByRole('switch', { name: /Voice input/ });
    expect(voiceSwitch.getAttribute('aria-checked')).toBe('true');

    fireEvent.click(voiceSwitch);

    expect(useCaretSettingsStore.getState().voiceEnabled).toBe(false);
    expect(voiceSwitch.getAttribute('aria-checked')).toBe('false');
  });
});
