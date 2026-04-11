import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ResultAsync } from 'neverthrow';

vi.mock('../lib/api', () => ({
  safe: {
    get: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: [] }), () => ({}) as any)),
    post: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    put: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    delete: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
  },
}));

import { bookService } from './bookService';
import { safe } from '../lib/api';

describe('bookService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // Appointments
  it('getAppointments calls GET /book/appointments with params', () => {
    const params = { status: 'confirmed', date: '2026-01-01' };
    bookService.getAppointments(params);
    expect(safe.get).toHaveBeenCalledWith('/book/appointments', { params });
  });

  it('getAppointment calls GET /book/appointments/:id', () => {
    bookService.getAppointment(3);
    expect(safe.get).toHaveBeenCalledWith('/book/appointments/3');
  });

  it('createAppointment calls POST /book/appointments', () => {
    const data = { service: 'Haircut', date: '2026-01-01' };
    bookService.createAppointment(data);
    expect(safe.post).toHaveBeenCalledWith('/book/appointments', data);
  });

  it('updateAppointment calls PUT /book/appointments/:id', () => {
    const data = { status: 'confirmed' };
    bookService.updateAppointment(5, data);
    expect(safe.put).toHaveBeenCalledWith('/book/appointments/5', data);
  });

  it('cancelAppointment calls PUT /book/appointments/:id/cancel', () => {
    bookService.cancelAppointment(5);
    expect(safe.put).toHaveBeenCalledWith('/book/appointments/5/cancel');
  });

  it('deleteAppointment calls DELETE /book/appointments/:id', () => {
    bookService.deleteAppointment(5);
    expect(safe.delete).toHaveBeenCalledWith('/book/appointments/5');
  });

  // Calendar
  it('getCalendarEvents calls GET /book/calendar with params', () => {
    const params = { start: '2026-01-01', end: '2026-01-31' };
    bookService.getCalendarEvents(params);
    expect(safe.get).toHaveBeenCalledWith('/book/calendar', { params });
  });

  // Dashboard
  it('getDashboardStats calls GET /book/dashboard/stats', () => {
    bookService.getDashboardStats();
    expect(safe.get).toHaveBeenCalledWith('/book/dashboard/stats');
  });

  it('getTodayAppointments calls GET /book/dashboard/today', () => {
    bookService.getTodayAppointments();
    expect(safe.get).toHaveBeenCalledWith('/book/dashboard/today');
  });

  // Services
  it('getServices calls GET /book/services with params', () => {
    const params = { search: 'cut' };
    bookService.getServices(params);
    expect(safe.get).toHaveBeenCalledWith('/book/services', { params });
  });

  it('createService calls POST /book/services', () => {
    const data = { name: 'Massage' };
    bookService.createService(data);
    expect(safe.post).toHaveBeenCalledWith('/book/services', data);
  });

  it('updateService calls PUT /book/services/:id', () => {
    bookService.updateService(2, { name: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/book/services/2', { name: 'Updated' });
  });

  it('deleteService calls DELETE /book/services/:id', () => {
    bookService.deleteService(2);
    expect(safe.delete).toHaveBeenCalledWith('/book/services/2');
  });

  // Resources
  it('getResources calls GET /book/resources with params', () => {
    const params = { type: 'room' };
    bookService.getResources(params);
    expect(safe.get).toHaveBeenCalledWith('/book/resources', { params });
  });

  it('createResource calls POST /book/resources', () => {
    const data = { name: 'Room A' };
    bookService.createResource(data);
    expect(safe.post).toHaveBeenCalledWith('/book/resources', data);
  });

  it('deleteResource calls DELETE /book/resources/:id', () => {
    bookService.deleteResource(4);
    expect(safe.delete).toHaveBeenCalledWith('/book/resources/4');
  });

  // Waitlist
  it('addToWaitlist calls POST /book/waitlist', () => {
    const data = { customer: 'Jane' };
    bookService.addToWaitlist(data);
    expect(safe.post).toHaveBeenCalledWith('/book/waitlist', data);
  });

  it('notifyWaitlist calls POST /book/waitlist/:id/notify', () => {
    bookService.notifyWaitlist(1);
    expect(safe.post).toHaveBeenCalledWith('/book/waitlist/1/notify');
  });

  it('removeFromWaitlist calls DELETE /book/waitlist/:id', () => {
    bookService.removeFromWaitlist(1);
    expect(safe.delete).toHaveBeenCalledWith('/book/waitlist/1');
  });

  // Customers
  it('getCustomers calls GET /book/customers', () => {
    bookService.getCustomers();
    expect(safe.get).toHaveBeenCalledWith('/book/customers', { params: undefined });
  });

  it('getCustomer calls GET /book/customers/:id', () => {
    bookService.getCustomer(9);
    expect(safe.get).toHaveBeenCalledWith('/book/customers/9');
  });

  // Settings
  it('getSettings calls GET /book/settings', () => {
    bookService.getSettings();
    expect(safe.get).toHaveBeenCalledWith('/book/settings');
  });

  it('updateSettings calls PUT /book/settings', () => {
    const data = { timezone: 'UTC' };
    bookService.updateSettings(data);
    expect(safe.put).toHaveBeenCalledWith('/book/settings', data);
  });
});
