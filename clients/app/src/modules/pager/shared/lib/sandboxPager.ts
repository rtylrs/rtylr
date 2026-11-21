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

// The server's alphabet and colour count: no 0, O, 1 or I.
const CODE_ALPHABET = '23456789ABCDEFGHJKLMNPQRSTUVWXYZ';
const COLORS = 6;

function colorOf(code: string): number {
  return (CODE_ALPHABET.indexOf(code[0] ?? '') + CODE_ALPHABET.indexOf(code[1] ?? '')) % COLORS;
}

/**
 * An in-memory stand-in for the pager API, for the public demo. It follows
 * the server's rules closely enough for the real board and phone screens to
 * behave as they do for a venue, and keeps everything in this page's memory:
 * nothing is sent anywhere and a reload starts over. Unlabelled tickets do not
 * expire here: the demo does not last fifteen minutes.
 */
export function createSandboxPager({
  venueName,
  now = () => new Date(),
  random = Math.random,
}: {
  /** Read on every answer, so the demo venue follows the page's language. */
  venueName: () => string;
  now?: () => Date;
  random?: () => number;
}): SandboxPager {
  const tickets: SandboxTicket[] = [];
  const settings: Required<GetSettingsResponse> = { enabled: true, labelMode: 'either' };
  let lastId = 0;

  const clock = (): string => now().toISOString();
  const isOpen = (ticket: SandboxTicket): boolean =>
    ticket.status === 'waiting' || ticket.status === 'ready';
  const staffView = ({ publicToken: _token, ...ticket }: SandboxTicket): Ticket => ({
    ...ticket,
    code: ticket.labelKind ? undefined : ticket.code,
  });

  const freeCode = (): string => {
    const used = new Set(tickets.filter((t) => isOpen(t) && !t.labelKind).map((t) => t.code));
    const free: string[] = [];
    for (const first of CODE_ALPHABET) {
      for (const second of CODE_ALPHABET) {
        if (!used.has(first + second)) free.push(first + second);
      }
    }
    const code = free[Math.floor(random() * free.length)];
    if (!code) throw new SandboxPagerError(429, PAGER_ERROR.tooManyOpenTickets);
    return code;
  };

  // The server's rule: an order number has a digit, anything else is a name.
  const parseLabel = (raw: string | undefined): { label: string; labelKind: string } => {
    const typed = (raw ?? '').trim();
    const number = normaliseTicketNumber(typed);
    const isNumber = /\d/.test(number) && /^[A-Za-z0-9٠-٩۰-۹]+$/.test(typed);
    if (settings.labelMode !== 'name' && isNumber && number.length === typed.length) {
      return { label: number, labelKind: 'number' };
    }
    const name = typed.split(/\s+/).filter(Boolean).join(' ');
    if (settings.labelMode === 'number' || !name || [...name].length > 40) {
      throw new SandboxPagerError(400, PAGER_ERROR.invalidLabel);
    }
    return { label: name, labelKind: 'name' };
  };

  const byToken = (token: string | undefined): SandboxTicket => {
    const ticket = tickets.find((candidate) => candidate.publicToken === token);
    if (!ticket || !settings.enabled) {
      throw new SandboxPagerError(404, PAGER_ERROR.ticketNotFound);
    }
    return ticket;
  };

  const byId = (id: string | undefined): SandboxTicket => {
    const ticket = tickets.find((candidate) => candidate.id === id);
    if (!ticket) throw new SandboxPagerError(404, PAGER_ERROR.ticketNotFound);
    return ticket;
  };

  const move = (
    id: string | undefined,
    from: string[],
    apply: (ticket: SandboxTicket) => void
  ): { ticket: Ticket } => {
    const ticket = byId(id);
    if (!from.includes(ticket.status ?? '')) {
      throw new SandboxPagerError(409, PAGER_ERROR.invalidTransition);
    }
    apply(ticket);
    return { ticket: staffView(ticket) };
  };

  return {
    scanTicket: async () => {
      if (!settings.enabled) throw new SandboxPagerError(404, PAGER_ERROR.venueNotFound);
      const code = freeCode();
      lastId += 1;
      const ticket: SandboxTicket = {
        id: String(lastId),
        code,
        color: colorOf(code),
        status: 'waiting',
        createdAt: clock(),
        publicToken: `sandbox-${lastId}`,
      };
      tickets.push(ticket);
      return { publicToken: ticket.publicToken, code, color: ticket.color, status: ticket.status };
    },

    getPublicTicket: async ({ publicToken }): Promise<GetPublicTicketResponse> => {
      const ticket = byToken(publicToken);
      const ahead =
        ticket.status === 'waiting' && ticket.labelKind
          ? tickets.filter(
              (other) =>
                other.status === 'waiting' &&
                other.labelKind &&
                Number(other.id) < Number(ticket.id)
            ).length
          : 0;
      return {
        label: ticket.label,
        labelKind: ticket.labelKind,
        code: ticket.labelKind ? undefined : ticket.code,
        color: ticket.color,
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

    labelTicket: async ({ id, label: raw }) => {
      const parsed = parseLabel(raw);
      const ticket = byId(id);
      if (ticket.status !== 'waiting') {
        throw new SandboxPagerError(409, PAGER_ERROR.invalidTransition);
      }
      const taken = tickets.some(
        (other) =>
          other !== ticket &&
          isOpen(other) &&
          other.labelKind === 'number' &&
          parsed.labelKind === 'number' &&
          other.label === parsed.label
      );
      if (taken) throw new SandboxPagerError(409, PAGER_ERROR.labelTaken);
      ticket.label = parsed.label;
      ticket.labelKind = parsed.labelKind;
      return { ticket: staffView(ticket) };
    },

    markReady: async ({ id }) => {
      if (!byId(id).labelKind) throw new SandboxPagerError(409, PAGER_ERROR.invalidTransition);
      return move(id, ['waiting'], (ticket) => {
        ticket.status = 'ready';
        ticket.readyAt = clock();
      });
    },

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

    updateSettings: async ({ enabled, labelMode }) => {
      if (enabled !== undefined) settings.enabled = enabled;
      if (labelMode !== undefined) settings.labelMode = labelMode;
      return { ...settings };
    },
  };
}
