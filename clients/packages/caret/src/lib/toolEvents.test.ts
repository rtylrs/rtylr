import { afterEach, describe, expect, it, vi } from 'vitest';
import { useCaretStore } from '../stores/caretStore';
import { onCaretToolSuccess } from './toolEvents';

describe('onCaretToolSuccess', () => {
  let stop = (): void => undefined;
  afterEach(() => stop());

  it('tells the host app once when a tool finishes', () => {
    const listener = vi.fn();
    stop = onCaretToolSuccess(listener);
    const store = useCaretStore.getState();
    store.appendBlockToLastAssistant({
      id: 'tool-1',
      type: 'tool_call',
      name: 'portal.customize',
      input: { id: 'caret' },
      status: 'running',
    });
    store.setToolStatus('tool-1', 'success', { output: { area: 'rail' } });
    store.setToolStatus('tool-1', 'success', { output: { area: 'rail' } });
    expect(listener).toHaveBeenCalledTimes(1);
    expect(listener).toHaveBeenCalledWith({
      name: 'portal.customize',
      input: { id: 'caret' },
      output: { area: 'rail' },
      approved: false,
    });
  });

  it('stays quiet for failures', () => {
    const listener = vi.fn();
    stop = onCaretToolSuccess(listener);
    const store = useCaretStore.getState();
    store.appendBlockToLastAssistant({
      id: 'tool-2',
      type: 'tool_call',
      name: 'crm.create_customer',
      input: {},
      status: 'running',
      requiresApproval: true,
    });
    store.setToolStatus('tool-2', 'error', { error: 'nope' });
    expect(listener).not.toHaveBeenCalled();
  });
});
