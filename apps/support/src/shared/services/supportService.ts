import { safe } from '../lib/api';

export interface TicketPayload {
  subject: string;
  body?: string;
  customerId?: number;
  channel?: string;
  priority?: string;
  assigneeId?: number;
  [key: string]: unknown;
}

export interface SupportSettingsPayload {
  businessHours?: Record<string, unknown>;
  autoReplyEnabled?: boolean;
  autoReplyMessage?: string;
  [key: string]: unknown;
}

export const supportService = {
  // Conversations / Inbox
  getConversations: (params?: { status?: string; assignee?: string; channel?: string }) =>
    safe.get('/support/conversations', { params }),
  getConversation: (id: number) => safe.get(`/support/conversations/${id}`),
  assignConversation: (id: number, assigneeId: number) =>
    safe.put(`/support/conversations/${id}/assign`, { assigneeId }),

  // Tickets
  getTickets: (params?: {
    status?: string;
    assignee?: string;
    search?: string;
    limit?: number;
    offset?: number;
  }) => safe.get('/support/tickets', { params }),
  getTicket: (id: number) => safe.get(`/support/tickets/${id}`),
  createTicket: (data: TicketPayload) => safe.post('/support/tickets', data),
  updateTicket: (id: number, data: Partial<TicketPayload>) =>
    safe.put(`/support/tickets/${id}`, data),
  updateTicketStatus: (id: number, status: string) =>
    safe.put(`/support/tickets/${id}/status`, { status }),

  // Ticket messages & notes
  getTicketMessages: (ticketId: number) => safe.get(`/support/tickets/${ticketId}/messages`),
  sendTicketMessage: (ticketId: number, data: { body: string; isInternal?: boolean }) =>
    safe.post(`/support/tickets/${ticketId}/messages`, data),

  // Customers
  getCustomers: (params?: { search?: string; limit?: number; offset?: number }) =>
    safe.get('/support/customers', { params }),
  getCustomer: (id: number) => safe.get(`/support/customers/${id}`),
  getCustomerTickets: (customerId: number) => safe.get(`/support/customers/${customerId}/tickets`),

  // Knowledge Base
  getArticles: (params?: { search?: string; category?: string }) =>
    safe.get('/support/knowledge/articles', { params }),
  getArticle: (id: number) => safe.get(`/support/knowledge/articles/${id}`),
  createArticle: (data: { title: string; body: string; category: string }) =>
    safe.post('/support/knowledge/articles', data),
  updateArticle: (id: number, data: { title?: string; body?: string; category?: string }) =>
    safe.put(`/support/knowledge/articles/${id}`, data),
  deleteArticle: (id: number) => safe.delete(`/support/knowledge/articles/${id}`),
  getArticleCategories: () => safe.get('/support/knowledge/categories'),

  // Canned Responses
  getCannedResponses: (params?: { search?: string }) =>
    safe.get('/support/canned-responses', { params }),
  getCannedResponse: (id: number) => safe.get(`/support/canned-responses/${id}`),
  createCannedResponse: (data: { title: string; body: string }) =>
    safe.post('/support/canned-responses', data),
  updateCannedResponse: (id: number, data: { title?: string; body?: string }) =>
    safe.put(`/support/canned-responses/${id}`, data),
  deleteCannedResponse: (id: number) => safe.delete(`/support/canned-responses/${id}`),

  // Settings
  getSettings: () => safe.get('/support/settings'),
  updateSettings: (data: SupportSettingsPayload) => safe.put('/support/settings', data),
};
