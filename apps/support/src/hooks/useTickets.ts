import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { supportService } from '../services/supportService';

export function useTickets(params?: { status?: string; assignee?: string; search?: string; limit?: number; offset?: number }) {
  return useQuery({
    queryKey: ['tickets', params],
    queryFn: () => unwrap(supportService.getTickets(params)),
  });
}

export function useTicket(id: number) {
  return useQuery({
    queryKey: ['ticket', id],
    queryFn: () => unwrap(supportService.getTicket(id)),
    enabled: !!id,
  });
}

export function useTicketMessages(ticketId: number) {
  return useQuery({
    queryKey: ['ticket-messages', ticketId],
    queryFn: () => unwrap(supportService.getTicketMessages(ticketId)),
    enabled: !!ticketId,
  });
}

export function useCreateTicket() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => unwrap(supportService.createTicket(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tickets'] });
    },
  });
}

export function useUpdateTicket() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) =>
      unwrap(supportService.updateTicket(id, data)),
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
      unwrap(supportService.updateTicketStatus(id, status)),
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
      unwrap(supportService.sendTicketMessage(ticketId, data)),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['ticket-messages', vars.ticketId] });
    },
  });
}
