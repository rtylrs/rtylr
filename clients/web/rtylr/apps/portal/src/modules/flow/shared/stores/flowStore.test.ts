import { beforeEach, describe, expect, it } from 'vitest';
import { useFlowStore } from './flowStore';

beforeEach(() =>
  useFlowStore.setState({
    members: [],
    projects: [],
    quickCreateOpen: false,
    quickCreateColumnId: undefined,
    projectCreateOpen: false,
    selectedTaskIds: [],
  })
);

describe('Flow interaction state', () => {
  it('keeps the selected column only while the task composer is open', () => {
    useFlowStore.getState().setQuickCreateOpen(true, 'column-7');
    expect(useFlowStore.getState().quickCreateColumnId).toBe('column-7');
    useFlowStore.getState().setQuickCreateOpen(false);
    expect(useFlowStore.getState().quickCreateColumnId).toBeUndefined();
    useFlowStore.getState().setQuickCreateOpen(true);
    expect(useFlowStore.getState().quickCreateColumnId).toBeUndefined();
  });

  it('toggles task selection and replaces or clears a batch', () => {
    useFlowStore.getState().toggleSelectTask('1');
    useFlowStore.getState().toggleSelectTask('2');
    useFlowStore.getState().toggleSelectTask('1');
    expect(useFlowStore.getState().selectedTaskIds).toEqual(['2']);
    useFlowStore.getState().selectAll(['3', '4']);
    expect(useFlowStore.getState().selectedTaskIds).toEqual(['3', '4']);
    useFlowStore.getState().clearSelection();
    expect(useFlowStore.getState().selectedTaskIds).toEqual([]);
  });
});
