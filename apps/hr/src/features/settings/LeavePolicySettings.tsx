import React, { useState } from 'react';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import { useCreateLeaveType } from '../../shared/hooks/queries/useLeave';
import { Button, Card, Input, Label, SectionHeader } from '../../shared/components/ui';

const LeavePolicySettings = (): React.JSX.Element => {
  const createLeaveType = useCreateLeaveType();
  const [name, setName] = useState('');
  const [maxCarryoverDays, setMaxCarryoverDays] = useState(0);
  const [minNoticeDays, setMinNoticeDays] = useState(1);
  const [isPaid, setIsPaid] = useState(true);
  const [requiresApproval, setRequiresApproval] = useState(true);

  const save = (event: React.FormEvent): void => {
    event.preventDefault();
    if (!name.trim() || maxCarryoverDays < 1 || minNoticeDays < 1) {
      toast.error('Enter a name and positive policy limits');
      return;
    }
    createLeaveType.mutate(
      {
        name: name.trim(),
        isPaid,
        requiresApproval,
        maxCarryoverDays,
        minNoticeDays,
      },
      {
        onSuccess: () => {
          setName('');
          setMaxCarryoverDays(0);
        },
      }
    );
  };

  return (
    <>
      <HelmetContainer title="Leave policy settings · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20, maxWidth: 560 }}>
        <SectionHeader
          title="Leave policies"
          subtitle="Create the leave types employees can request."
        />
        <Card>
          <form onSubmit={save} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <div>
              <Label>Name</Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Annual Leave"
                required
              />
            </div>
            <div>
              <Label>Maximum carryover days</Label>
              <Input
                type="number"
                min={1}
                value={maxCarryoverDays}
                onChange={(e) => setMaxCarryoverDays(Number(e.target.value))}
                required
              />
            </div>
            <div>
              <Label>Minimum notice days</Label>
              <Input
                type="number"
                min={1}
                value={minNoticeDays}
                onChange={(e) => setMinNoticeDays(Number(e.target.value))}
                required
              />
            </div>
            <label>
              <input
                type="checkbox"
                checked={isPaid}
                onChange={(e) => setIsPaid(e.target.checked)}
              />{' '}
              Paid leave
            </label>
            <label>
              <input
                type="checkbox"
                checked={requiresApproval}
                onChange={(e) => setRequiresApproval(e.target.checked)}
              />{' '}
              Requires manager approval
            </label>
            <Button type="submit" variant="primary" disabled={createLeaveType.isPending}>
              {createLeaveType.isPending ? 'Creating…' : 'Create leave policy'}
            </Button>
          </form>
        </Card>
      </div>
    </>
  );
};

export default LeavePolicySettings;
