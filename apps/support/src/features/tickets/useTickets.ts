import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { supportService } from '../../shared/services/supportService';
import { ticketSchema, ticketMessageSchema } from '../../shared/schemas/tickets';

export function useTickets(params?: {
  status?: string;
  assignee?: string;
  search?: string;
  limit?: number;
  offset?: number;
}) {
  return useQuery({
    queryKey: ['tickets', params],
    queryFn: async () => {
      const res = await unwrap(supportService.getTickets(params));
      const raw = res ?? [];
      return z.array(ticketSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useTicket(id: number) {
  return useQuery({
    queryKey: ['ticket', id],
    queryFn: async () => {
      const res = await unwrap(supportService.getTicket(id));
      return ticketSchema.parse(res);
    },
    enabled: !!id,
  });
}

export function useTicketMessages(ticketId: number) {
  return useQuery({
    queryKey: ['ticket-messages', ticketId],
    queryFn: async () => {
      const res = await unwrap(supportService.getTicketMessages(ticketId));
      const raw = res ?? [];
      return z.array(ticketMessageSchema).parse(Array.isArray(raw) ? raw : []);
    },
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
    mutationFn: ({
      ticketId,
      data,
    }: {
      ticketId: number;
      data: { body: string; isInternal?: boolean };
    }) => unwrap(supportService.sendTicketMessage(ticketId, data)),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['ticket-messages', vars.ticketId] });
    },
  });
}
