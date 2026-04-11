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

import { supportService } from './supportService';
import { safe } from '../lib/api';

describe('supportService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // Conversations
  it('getConversations calls GET /support/conversations with params', () => {
    const params = { status: 'open', channel: 'email' };
    supportService.getConversations(params);
    expect(safe.get).toHaveBeenCalledWith('/support/conversations', { params });
  });

  it('getConversation calls GET /support/conversations/:id', () => {
    supportService.getConversation(1);
    expect(safe.get).toHaveBeenCalledWith('/support/conversations/1');
  });

  it('assignConversation calls PUT /support/conversations/:id/assign', () => {
    supportService.assignConversation(1, 5);
    expect(safe.put).toHaveBeenCalledWith('/support/conversations/1/assign', { assigneeId: 5 });
  });

  // Tickets
  it('getTickets calls GET /support/tickets with params', () => {
    const params = { status: 'open', search: 'bug' };
    supportService.getTickets(params);
    expect(safe.get).toHaveBeenCalledWith('/support/tickets', { params });
  });

  it('createTicket calls POST /support/tickets', () => {
    const data = { subject: 'Issue', body: 'Help' };
    supportService.createTicket(data);
    expect(safe.post).toHaveBeenCalledWith('/support/tickets', data);
  });

  it('updateTicket calls PUT /support/tickets/:id', () => {
    supportService.updateTicket(2, { priority: 'high' });
    expect(safe.put).toHaveBeenCalledWith('/support/tickets/2', { priority: 'high' });
  });

  it('updateTicketStatus calls PUT /support/tickets/:id/status', () => {
    supportService.updateTicketStatus(2, 'closed');
    expect(safe.put).toHaveBeenCalledWith('/support/tickets/2/status', { status: 'closed' });
  });

  // Ticket messages
  it('getTicketMessages calls GET /support/tickets/:id/messages', () => {
    supportService.getTicketMessages(3);
    expect(safe.get).toHaveBeenCalledWith('/support/tickets/3/messages');
  });

  it('sendTicketMessage calls POST /support/tickets/:id/messages', () => {
    const data = { body: 'Reply here', isInternal: false };
    supportService.sendTicketMessage(3, data);
    expect(safe.post).toHaveBeenCalledWith('/support/tickets/3/messages', data);
  });

  // Customers
  it('getCustomers calls GET /support/customers', () => {
    supportService.getCustomers({ search: 'john' });
    expect(safe.get).toHaveBeenCalledWith('/support/customers', { params: { search: 'john' } });
  });

  it('getCustomerTickets calls GET /support/customers/:id/tickets', () => {
    supportService.getCustomerTickets(4);
    expect(safe.get).toHaveBeenCalledWith('/support/customers/4/tickets');
  });

  // Knowledge Base
  it('getArticles calls GET /support/knowledge/articles', () => {
    supportService.getArticles({ category: 'faq' });
    expect(safe.get).toHaveBeenCalledWith('/support/knowledge/articles', {
      params: { category: 'faq' },
    });
  });

  it('createArticle calls POST /support/knowledge/articles', () => {
    const data = { title: 'How to', body: 'Steps', category: 'faq' };
    supportService.createArticle(data);
    expect(safe.post).toHaveBeenCalledWith('/support/knowledge/articles', data);
  });

  it('deleteArticle calls DELETE /support/knowledge/articles/:id', () => {
    supportService.deleteArticle(7);
    expect(safe.delete).toHaveBeenCalledWith('/support/knowledge/articles/7');
  });

  it('getArticleCategories calls GET /support/knowledge/categories', () => {
    supportService.getArticleCategories();
    expect(safe.get).toHaveBeenCalledWith('/support/knowledge/categories');
  });

  // Canned Responses
  it('createCannedResponse calls POST /support/canned-responses', () => {
    const data = { title: 'Thanks', body: 'Thank you!' };
    supportService.createCannedResponse(data);
    expect(safe.post).toHaveBeenCalledWith('/support/canned-responses', data);
  });

  it('deleteCannedResponse calls DELETE /support/canned-responses/:id', () => {
    supportService.deleteCannedResponse(2);
    expect(safe.delete).toHaveBeenCalledWith('/support/canned-responses/2');
  });

  // Settings
  it('getSettings calls GET /support/settings', () => {
    supportService.getSettings();
    expect(safe.get).toHaveBeenCalledWith('/support/settings');
  });

  it('updateSettings calls PUT /support/settings', () => {
    const data = { autoAssign: true };
    supportService.updateSettings(data);
    expect(safe.put).toHaveBeenCalledWith('/support/settings', data);
  });
});
