export type EmploymentType = 'full_time' | 'part_time' | 'contractor';
export type ShiftStatus = 'scheduled' | 'in_progress' | 'completed' | 'absent';

export interface Staff {
  id: string;
  name: string;
  email: string;
  phone?: string;
  role: string;
  department?: string;
  employmentType: EmploymentType;
  salary?: number;
  hourlyRate?: number;
  startDate: string;
  apps?: string[];
}

export interface Shift {
  id: string;
  staffId: string;
  date: string;
  startTime: string;
  endTime: string;
  status: ShiftStatus;
  hoursWorked?: number;
  notes?: string;
}

export interface PayrollRecord {
  id: string;
  staffId: string;
  period: string;
  basePay: number;
  overtime?: number;
  deductions?: number;
  netPay: number;
  paidAt?: string;
}
