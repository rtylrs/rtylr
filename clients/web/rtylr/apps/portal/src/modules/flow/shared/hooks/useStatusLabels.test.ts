import { describe, expect, it, vi } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useStatusLabels } from './useStatusLabels';
import { STATUS_LABEL_KEYS } from '../lib/constantsFlow';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

describe('useStatusLabels', () => {
  it('resolves every status through the shared STATUS_LABEL_KEYS map', () => {
    const { result } = renderHook(() => useStatusLabels());

    for (const status of Object.keys(STATUS_LABEL_KEYS) as (keyof typeof STATUS_LABEL_KEYS)[]) {
      expect(result.current[status]).toBe(STATUS_LABEL_KEYS[status]);
    }
  });

  it("resolves 'todo' to the flow.taskDetail.todo key, the same one the task detail status picker uses", () => {
    const { result } = renderHook(() => useStatusLabels());

    expect(result.current.todo).toBe('flow.taskDetail.todo');
  });
});
