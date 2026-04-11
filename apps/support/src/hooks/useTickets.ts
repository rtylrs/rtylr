import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { supportService } from '../services/supportService';

export function useTickets(params?: { status?: string; assignee?: string; search?: string; limit?: number; offset?: number }) {
  return useQuery({
    queryKey: ['tickets', params],
    queryFn: () => supportService.getTickets(params).then((r) => r.data),
  });
}

export function useTicket(id: number) {
  return useQuery({
    queryKey: ['ticket', id],
    queryFn: () => supportService.getTicket(id).then((r) => r.data),
    enabled: !!id,
  });
}

export function useTicketMessages(ticketId: number) {
  return useQuery({
    queryKey: ['ticket-messages', ticketId],
    queryFn: () => supportService.getTicketMessages(ticketId).then((r) => r.data),
    enabled: !!ticketId,
  });
}

export function useCreateTicket() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => supportService.createTicket(data).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tickets'] });
    },
  });
}

export function useUpdateTicket() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) =>
      supportService.updateTicket(id, data).then((r) => r.data),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['tickets'] });
      qc.invalidateQueries({ queryKey: ['ticket', vars.id] });
    },
  });
}

export function useUpdateTicketStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: number; status: string }) =>
      supportService.updateTicketStatus(id, status).then((r) => r.data),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['tickets'] });
      qc.invalidateQueries({ queryKey: ['ticket', vars.id] });
    },
  });
}

export function useSendTicketMessage() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ ticketId, data }: { ticketId: number; data: { body: string; isInternal?: boolean } }) =>
      supportService.sendTicketMessage(ticketId, data).then((r) => r.data),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['ticket-messages', vars.ticketId] });
    },
  });
}
