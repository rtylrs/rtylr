import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { bookService } from '../services/bookService';

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
      const res = await bookService.getAppointments(filters);
      return res.data?.data ?? res.data ?? [];
    },
  });
}

export function useAppointment(id: number | undefined) {
  return useQuery({
    queryKey: ['appointment', id],
    enabled: !!id,
    queryFn: async () => {
      if (!id) return null;
      const res = await bookService.getAppointment(id);
      return res.data?.data ?? res.data ?? null;
    },
  });
}

export function useCreateAppointment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => bookService.createAppointment(data),
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
    mutationFn: ({ id, data }: { id: number; data: any }) => bookService.updateAppointment(id, data),
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
    mutationFn: (id: number) => bookService.cancelAppointment(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['appointments'] });
      qc.invalidateQueries({ queryKey: ['calendar'] });
      qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
}
