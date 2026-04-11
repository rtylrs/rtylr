import api from '../lib/api';

export const supportService = {
  // Conversations / Inbox
  getConversations: (params?: { status?: string; assignee?: string; channel?: string }) =>
    api.get('/support/conversations', { params }),
  getConversation: (id: number) => api.get(`/support/conversations/${id}`),
  assignConversation: (id: number, assigneeId: number) =>
    api.put(`/support/conversations/${id}/assign`, { assigneeId }),

  // Tickets
  getTickets: (params?: { status?: string; assignee?: string; search?: string; limit?: number; offset?: number }) =>
    api.get('/support/tickets', { params }),
  getTicket: (id: number) => api.get(`/support/tickets/${id}`),
  createTicket: (data: any) => api.post('/support/tickets', data),
  updateTicket: (id: number, data: any) => api.put(`/support/tickets/${id}`, data),
  updateTicketStatus: (id: number, status: string) =>
    api.put(`/support/tickets/${id}/status`, { status }),

  // Ticket messages & notes
  getTicketMessages: (ticketId: number) => api.get(`/support/tickets/${ticketId}/messages`),
  sendTicketMessage: (ticketId: number, data: { body: string; isInternal?: boolean }) =>
    api.post(`/support/tickets/${ticketId}/messages`, data),

  // Customers
  getCustomers: (params?: { search?: string; limit?: number; offset?: number }) =>
    api.get('/support/customers', { params }),
  getCustomer: (id: number) => api.get(`/support/customers/${id}`),
  getCustomerTickets: (customerId: number) => api.get(`/support/customers/${customerId}/tickets`),

  // Knowledge Base
  getArticles: (params?: { search?: string; category?: string }) =>
    api.get('/support/knowledge/articles', { params }),
  getArticle: (id: number) => api.get(`/support/knowledge/articles/${id}`),
  createArticle: (data: { title: string; body: string; category: string }) =>
    api.post('/support/knowledge/articles', data),
  updateArticle: (id: number, data: { title?: string; body?: string; category?: string }) =>
    api.put(`/support/knowledge/articles/${id}`, data),
  deleteArticle: (id: number) => api.delete(`/support/knowledge/articles/${id}`),
  getArticleCategories: () => api.get('/support/knowledge/categories'),

  // Canned Responses
  getCannedResponses: (params?: { search?: string }) =>
    api.get('/support/canned-responses', { params }),
  getCannedResponse: (id: number) => api.get(`/support/canned-responses/${id}`),
  createCannedResponse: (data: { title: string; body: string }) =>
    api.post('/support/canned-responses', data),
  updateCannedResponse: (id: number, data: { title?: string; body?: string }) =>
    api.put(`/support/canned-responses/${id}`, data),
  deleteCannedResponse: (id: number) => api.delete(`/support/canned-responses/${id}`),

  // Settings
  getSettings: () => api.get('/support/settings'),
  updateSettings: (data: any) => api.put('/support/settings', data),
};
