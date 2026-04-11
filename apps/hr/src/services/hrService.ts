import { safe } from '../lib/api';

export const hrService = {
  // Staff
  getStaff: () => safe.get('/hr/staff'),
  createStaff: (data: any) => safe.post('/hr/staff', data),
  updateStaff: (id: number, data: any) => safe.put(`/hr/staff/${id}`, data),
  deleteStaff: (id: number) => safe.delete(`/hr/staff/${id}`),
  getDashboard: () => safe.get('/hr/dashboard'),

  // Departments
  getDepartments: () => safe.get('/hr/departments'),
  createDepartment: (data: any) => safe.post('/hr/departments', data),
  updateDepartment: (id: number, data: any) => safe.put(`/hr/departments/${id}`, data),

  // Shifts
  getShifts: () => safe.get('/hr/shifts'),
  createShift: (data: any) => safe.post('/hr/shifts', data),
  clockIn: (data: any) => safe.post('/hr/shifts/clock-in', data),
  clockOut: (data: any) => safe.post('/hr/shifts/clock-out', data),

  // Time Off
  getTimeOff: () => safe.get('/hr/time-off'),
  submitTimeOff: (data: any) => safe.post('/hr/time-off', data),
  approveTimeOff: (data: any) => safe.put('/hr/time-off/approve', data),

  // Payroll
  getPayroll: () => safe.get('/hr/payroll'),
  runPayroll: (data: any) => safe.post('/hr/payroll/run', data),
  markPayrollPaid: (data: any) => safe.put('/hr/payroll/mark-paid', data),
};
