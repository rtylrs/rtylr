import React, { useState } from 'react';
import { useNavigate, Link } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { toast } from 'sonner';
import { useStaff } from '../staff/useStaff';
import { useSubmitTimeOff } from '../../shared/hooks/useHrExtras';
import {
  Card,
  SectionHeader,
  Button,
  Label,
  Input,
  Textarea,
  Select,
  Breadcrumbs,
  useIsMobile,
} from '../../shared/components/ui';

const TYPES = ['Vacation', 'Sick', 'Personal', 'Bereavement', 'Unpaid'] as const;

const TimeOffNew: React.FC = () => {
  const isMobile = useIsMobile(900);
  const nav = useNavigate();
  const { data: staff } = useStaff();
  const submit = useSubmitTimeOff();

  const [staffId, setStaffId] = useState<number>(0);
  const [type, setType] = useState<(typeof TYPES)[number]>('Vacation');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [reason, setReason] = useState('');

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const who = (staff ?? []).find((s) => s.id === staffId);
    if (!who) {
      toast.error('Select an employee');
      return;
    }
    if (!startDate || !endDate) {
      toast.error('Select a date range');
      return;
    }
    if (endDate < startDate) {
      toast.error('End date cannot be before start date');
      return;
    }
    submit.mutate(
      { staffId, staffName: who.name, type, startDate, endDate, reason },
      {
        onSuccess: () => {
          toast.success('Time off request submitted');
          nav({ to: '/time-off' });
        },
        onError: () => toast.error('Failed to submit'),
      }
    );
  };

  return (
    <>
      <HelmetContainer title="New time off request · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16, maxWidth: 600 }}>
        <Breadcrumbs items={[{ label: 'Time Off', href: '/time-off' }, { label: 'New request' }]} />
        <SectionHeader title="Request Time Off" subtitle="Submit a new PTO request for review" />
        <Card>
          <form onSubmit={onSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <div>
              <Label>Employee</Label>
              <Select value={staffId} onChange={(e) => setStaffId(Number(e.target.value))}>
                <option value={0}>Select employee</option>
                {(staff ?? []).map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name} ({s.role})
                  </option>
                ))}
              </Select>
            </div>
            <div>
              <Label>Type</Label>
              <Select
                value={type}
                onChange={(e) => setType(e.target.value as (typeof TYPES)[number])}
              >
                {TYPES.map((t) => (
                  <option key={t} value={t}>
                    {t}
                  </option>
                ))}
              </Select>
            </div>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr',
                gap: 10,
              }}
            >
              <div>
                <Label>Start date</Label>
                <Input
                  type="date"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                />
              </div>
              <div>
                <Label>End date</Label>
                <Input type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} />
              </div>
            </div>
            <div>
              <Label>Reason</Label>
              <Textarea
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder="Brief reason for request"
              />
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', flexWrap: 'wrap' }}>
              <Link to="/time-off">
                <Button type="button" variant="ghost">
                  Cancel
                </Button>
              </Link>
              <Button type="submit" variant="primary" disabled={submit.isPending}>
                {submit.isPending ? 'Submitting...' : 'Submit request'}
              </Button>
            </div>
          </form>
        </Card>
      </div>
    </>
  );
};

export default TimeOffNew;
