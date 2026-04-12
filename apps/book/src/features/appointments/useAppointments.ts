import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import type {
  CreateAppointmentRequest,
  UpdateAppointmentRequest,
} from '@voxire/contracts/book/v1/appointments';
import { bookService } from '../../shared/services/bookService';

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
    queryFn: () => unwrap(bookService.getAppointments(filters)),
  });
}

export function useAppointment(id: number | undefined) {
  return useQuery({
    queryKey: ['appointment', id],
    enabled: !!id,
    queryFn: () => unwrap(bookService.getAppointment(id!)),
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
