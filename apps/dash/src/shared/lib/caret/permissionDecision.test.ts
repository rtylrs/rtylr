import { describe, expect, it, beforeEach } from 'vitest';
import { resolveDecision } from './permissionDecision';
import { rememberApproval, clearApprovals } from './approvals';

beforeEach(() => clearApprovals('c1'));

describe('resolveDecision', () => {
  it('auto-approves read tools when allowed', () => {
    expect(
      resolveDecision({
        chatId: 'c1',
        tool: 'pos.list_orders',
        scope: 'read',
        autoApproveReadOnly: true,
      })
    ).toBe('approved');
  });

  it('returns pending for write tools by default', () => {
    expect(
      resolveDecision({
        chatId: 'c1',
        tool: 'pos.refund_order',
        scope: 'write',
        autoApproveReadOnly: true,
      })
    ).toBe('pending');
  });

  it('respects cached user decisions', () => {
    rememberApproval('c1', 'pos.refund_order', 'denied');
    expect(
      resolveDecision({
        chatId: 'c1',
        tool: 'pos.refund_order',
        scope: 'write',
        autoApproveReadOnly: true,
      })
    ).toBe('denied');
  });
});
