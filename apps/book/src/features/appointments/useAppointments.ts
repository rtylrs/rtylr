import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import type {
  CreateAppointmentRequest,
  UpdateAppointmentRequest,
} from '@voxire/contracts/book/v1/appointments';
import { bookService } from '../../shared/services/bookService';
import { appointmentSchema } from '../../shared/schemas/appointments';

export interface AppointmentFilters {
  search?: string;
  status?: string;
  date?: string;
  limit?: number;
  offset?: number;
}

export function useAppointments(filters?: AppointmentFilters) {
  return useQuery({
    queryKey: ['appointments', filters],
    queryFn: async () => {
      const res = await unwrap(bookService.getAppointments(filters));
      const raw = (res as any)?.appointments ?? res;
      return z.array(appointmentSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useAppointment(id: number | undefined) {
  return useQuery({
    queryKey: ['appointment', id],
    enabled: !!id,
    queryFn: async () => {
      const res = await unwrap(bookService.getAppointment(id!));
      return appointmentSchema.parse(res);
    },
  });
}

export function useCreateAppointment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateAppointmentRequest) => unwrap(bookService.createAppointment(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['appointments'] });
      qc.invalidateQueries({ queryKey: ['calendar'] });
      qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
}

export function useUpdateAppointment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateAppointmentRequest }) =>
      unwrap(bookService.updateAppointment(id, data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['appointments'] });
      qc.invalidateQueries({ queryKey: ['calendar'] });
      qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
}

export function useCancelAppointment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(bookService.cancelAppointment(id)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['appointments'] });
      qc.invalidateQueries({ queryKey: ['calendar'] });
      qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
}
