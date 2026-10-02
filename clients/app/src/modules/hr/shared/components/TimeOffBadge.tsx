import React from 'react';
import { useTranslation } from 'react-i18next';
import { Pill } from './ui';

type Status = 'Pending' | 'Approved' | 'Rejected';

const COLORS: Record<string, string> = {
  Pending: '#f59e0b',
  Approved: '#10b981',
  Rejected: '#ef4444',
  pending: '#f59e0b',
  approved: '#10b981',
  rejected: '#ef4444',
};

interface Props {
  status: Status | string;
}

const TimeOffBadge: React.FC<Props> = ({ status }) => {
  const { t } = useTranslation('hrCommon');
  const lower = (status || '').toLowerCase();
  const label = t(`hrLeave.statuses.${lower}`, { defaultValue: status });
  return <Pill color={COLORS[status] || COLORS[lower] || '#94a3b8'}>{label}</Pill>;
};

export default TimeOffBadge;
