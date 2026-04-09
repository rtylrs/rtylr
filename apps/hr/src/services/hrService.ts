import api from '../lib/api';

export const hrService = {
  // Staff
  getStaff: () => api.get('/hr/staff'),
  createStaff: (data: any) => api.post('/hr/staff', data),
  updateStaff: (id: number, data: any) => api.put(`/hr/staff/${id}`, data),
  deleteStaff: (id: number) => api.delete(`/hr/staff/${id}`),
  getDashboard: () => api.get('/hr/dashboard'),

  // Departments
  getDepartments: () => api.get('/hr/departments'),
  createDepartment: (data: any) => api.post('/hr/departments', data),
  updateDepartment: (id: number, data: any) => api.put(`/hr/departments/${id}`, data),

  // Shifts
  getShifts: () => api.get('/hr/shifts'),
  createShift: (data: any) => api.post('/hr/shifts', data),
  clockIn: (data: any) => api.post('/hr/shifts/clock-in', data),
  clockOut: (data: any) => api.post('/hr/shifts/clock-out', data),

  // Time Off
  getTimeOff: () => api.get('/hr/time-off'),
  submitTimeOff: (data: any) => api.post('/hr/time-off', data),
  approveTimeOff: (data: any) => api.put('/hr/time-off/approve', data),

  // Payroll
  getPayroll: () => api.get('/hr/payroll'),
  runPayroll: (data: any) => api.post('/hr/payroll/run', data),
  markPayrollPaid: (data: any) => api.put('/hr/payroll/mark-paid', data),
};
