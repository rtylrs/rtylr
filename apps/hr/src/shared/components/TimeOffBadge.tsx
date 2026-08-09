import React from 'react';
import { Pill } from './ui';

type Status = 'Pending' | 'Approved' | 'Rejected';

const COLORS: Record<Status, string> = {
  Pending: '#f59e0b',
  Approved: '#10b981',
  Rejected: '#ef4444',
};

interface Props {
  status: Status;
}

const TimeOffBadge: React.FC<Props> = ({ status }) => <Pill color={COLORS[status]}>{status}</Pill>;

export default TimeOffBadge;
