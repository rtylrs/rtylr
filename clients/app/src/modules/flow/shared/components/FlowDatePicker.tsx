import React from 'react';
import { DatePicker, type DatePickerProps } from '@voxire/ui';

import { useTimeSettingsStore } from '../stores/timeSettingsStore';

/**
 * The shared date picker with weeks starting on the day chosen in Flow's time
 * settings, so every picker agrees with the Calendar views.
 */
const FlowDatePicker: React.FC<DatePickerProps> = (props) => {
  const weekStart = useTimeSettingsStore((s) => s.weekStart);
  return <DatePicker weekStartsOn={weekStart === 'sunday' ? 0 : 1} {...props} />;
};

export default FlowDatePicker;
