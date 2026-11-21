import type {
  GetPublicTicketResponse,
  GetSettingsResponse,
  Ticket,
} from '@voxire/voxie/pager/v1/types';
import { PAGER_ERROR } from './apiError';
import type { PublicPagerClient, StaffPagerClient } from './pagerClients';
import { normaliseTicketNumber } from './ticketNumber';

/** Thrown with the same `statusCode` and `code` the pager API sends. */
export class SandboxPagerError extends Error {
  constructor(
    readonly statusCode: number,
    readonly code: string
  ) {
    super(code);
    this.name = 'SandboxPagerError';
  }
}

interface SandboxTicket extends Ticket {
  publicToken: string;
}

export type SandboxPager = PublicPagerClient & StaffPagerClient;

/**
 * An in-memory stand-in for the pager API, for the public demo. It follows
 * the server's rules closely enough for the real board and phone screens to
 * behave as they do for a venue, and keeps everything in this page's memory:
 * nothing is sent anywhere and a reload starts over.
 */
export function createSandboxPager({
  venueName,
  now = () => new Date(),
}: {
  /** Read on every answer, so the demo venue follows the page's language. */
  venueName: () => string;
  now?: () => Date;
}): SandboxPager {
  const tickets: SandboxTicket[] = [];
  const settings: Required<GetSettingsResponse> = { enabled: true, strict: false };
  let lastId = 0;

  const clock = (): string => now().toISOString();
  const isOpen = (ticket: SandboxTicket): boolean =>
    ticket.status === 'waiting' || ticket.status === 'ready';
  const staffView = ({ publicToken: _token, ...ticket }: SandboxTicket): Ticket => ({ ...ticket });

  const requireVenue = (): void => {
    if (!settings.enabled) throw new SandboxPagerError(404, PAGER_ERROR.venueNotFound);
  };

  const numberFrom = (raw: string | undefined): string => {
    const number = normaliseTicketNumber(raw ?? '');
    if (!number) throw new SandboxPagerError(400, PAGER_ERROR.invalidNumber);
    return number;
  };

  const open = (number: string, source: 'staff' | 'customer', create: boolean): SandboxTicket => {
    const existing = tickets.find((ticket) => isOpen(ticket) && ticket.number === number);
    if (existing) {
      if (source === 'customer') existing.phoneLinked = true;
      return existing;
    }
    if (!create) throw new SandboxPagerError(404, PAGER_ERROR.ticketNotFound);
    lastId += 1;
    const ticket: SandboxTicket = {
      id: String(lastId),
      number,
      status: 'waiting',
      source,
      phoneLinked: source === 'customer',
      createdAt: clock(),
      publicToken: `sandbox-${lastId}`,
    };
    tickets.push(ticket);
    return ticket;
  };

  const byToken = (token: string | undefined): SandboxTicket => {
    const ticket = tickets.find((candidate) => candidate.publicToken === token);
    if (!ticket || !settings.enabled) {
      throw new SandboxPagerError(404, PAGER_ERROR.ticketNotFound);
    }
    return ticket;
  };

  const move = (
    id: string | undefined,
    from: string[],
    apply: (ticket: SandboxTicket) => void
  ): { ticket: Ticket } => {
    const ticket = tickets.find((candidate) => candidate.id === id);
    if (!ticket) throw new SandboxPagerError(404, PAGER_ERROR.ticketNotFound);
    if (!from.includes(ticket.status ?? '')) {
      throw new SandboxPagerError(409, PAGER_ERROR.invalidTransition);
    }
    apply(ticket);
    return { ticket: staffView(ticket) };
  };

  return {
    getPublicVenue: async () => {
      requireVenue();
      return { name: venueName(), ...settings };
    },

    joinTicket: async ({ number }) => {
      const normalised = numberFrom(number);
      requireVenue();
      const ticket = open(normalised, 'customer', !settings.strict);
      return { publicToken: ticket.publicToken, number: ticket.number, status: ticket.status };
    },

    getPublicTicket: async ({ publicToken }): Promise<GetPublicTicketResponse> => {
      const ticket = byToken(publicToken);
      const ahead =
        ticket.status === 'waiting'
          ? tickets.filter(
              (other) => other.status === 'waiting' && Number(other.id) < Number(ticket.id)
            ).length
          : 0;
      return {
        number: ticket.number,
        status: ticket.status,
        venueName: venueName(),
        ahead,
        readyAt: ticket.readyAt,
        remindedAt: ticket.remindedAt,
        acknowledgedAt: ticket.acknowledgedAt,
        serverTime: clock(),
      };
    },

    acknowledgeTicket: async ({ publicToken }) => {
      const ticket = byToken(publicToken);
      if (ticket.status === 'ready' && !ticket.acknowledgedAt) ticket.acknowledgedAt = clock();
      return {};
    },

    listTickets: async () => ({
      tickets: tickets.filter(isOpen).map(staffView),
      serverTime: clock(),
    }),

    createTicket: async ({ number }) => ({
      ticket: staffView(open(numberFrom(number), 'staff', true)),
    }),

    markReady: async ({ id }) =>
      move(id, ['waiting'], (ticket) => {
        ticket.status = 'ready';
        ticket.readyAt = clock();
      }),

    remindTicket: async ({ id }) =>
      move(id, ['ready'], (ticket) => {
        ticket.remindedAt = clock();
        ticket.acknowledgedAt = undefined;
      }),

    undoReady: async ({ id }) =>
      move(id, ['ready'], (ticket) => {
        ticket.status = 'waiting';
        ticket.readyAt = undefined;
        ticket.remindedAt = undefined;
        ticket.acknowledgedAt = undefined;
      }),

    collectTicket: async ({ id }) =>
      move(id, ['waiting', 'ready'], (ticket) => {
        ticket.status = 'collected';
      }),

    getSettings: async () => ({ ...settings }),

    updateSettings: async ({ enabled, strict }) => {
      if (enabled !== undefined) settings.enabled = enabled;
      if (strict !== undefined) settings.strict = strict;
      return { ...settings };
    },
  };
}
