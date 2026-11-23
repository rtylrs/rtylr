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

// The server's match numbers: three digits while any is free, then four.
const MATCH_NUMBER_SPANS: readonly (readonly [number, number])[] = [
  [100, 999],
  [1000, 9999],
];

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
  const settings: Required<GetSettingsResponse> = {
    enabled: true,
    labelMode: 'either',
    reviewUrl: '',
    prepMinutes: 0,
    customerNumbers: false,
    ringtone: 'bell',
  };
  let lastId = 0;

  const clock = (): string => now().toISOString();
  // As the server: an estimate from now, none for 0 minutes.
  const expectedAt = (minutes: number): string | undefined =>
    minutes > 0 ? new Date(now().getTime() + minutes * 60_000).toISOString() : undefined;
  const isOpen = (ticket: SandboxTicket): boolean =>
    ticket.status === 'waiting' || ticket.status === 'ready';
  const staffView = ({ publicToken: _token, ...ticket }: SandboxTicket): Ticket => ({
    ...ticket,
    code: ticket.labelKind ? undefined : ticket.code,
  });

  // As the server: a random number no open unlabelled ticket shows.
  const freeCode = (): string => {
    const used = new Set(tickets.filter((t) => isOpen(t) && !t.labelKind).map((t) => t.code));
    for (const [min, max] of MATCH_NUMBER_SPANS) {
      const free: string[] = [];
      for (let n = min; n <= max; n += 1) {
        if (!used.has(String(n))) free.push(String(n));
      }
      const code = free[Math.floor(random() * free.length)];
      if (code) return code;
    }
    throw new SandboxPagerError(429, PAGER_ERROR.tooManyOpenTickets);
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
    getPublicVenue: async () => {
      if (!settings.enabled) throw new SandboxPagerError(404, PAGER_ERROR.venueNotFound);
      return {
        name: venueName(),
        enabled: true,
        customerNumbers: settings.customerNumbers,
      };
    },

    joinTicket: async ({ number: raw }) => {
      if (!settings.enabled) throw new SandboxPagerError(404, PAGER_ERROR.venueNotFound);
      if (!settings.customerNumbers) {
        throw new SandboxPagerError(409, PAGER_ERROR.joinByNumberOff);
      }
      const number = normaliseTicketNumber((raw ?? '').trim());
      if (!/\d/.test(number) || !/^[A-Z0-9]{1,12}$/.test(number)) {
        throw new SandboxPagerError(400, PAGER_ERROR.invalidLabel);
      }
      const taken = tickets.some(
        (other) => isOpen(other) && other.labelKind === 'number' && other.label === number
      );
      if (taken) throw new SandboxPagerError(409, PAGER_ERROR.labelTaken);
      lastId += 1;
      const ticket: SandboxTicket = {
        id: String(lastId),
        label: number,
        labelKind: 'number',
        status: 'waiting',
        createdAt: clock(),
        expectedReadyAt: expectedAt(settings.prepMinutes),
        publicToken: `sandbox-${lastId}`,
      };
      tickets.push(ticket);
      return { publicToken: ticket.publicToken, label: number, status: ticket.status };
    },

    scanTicket: async () => {
      if (!settings.enabled) throw new SandboxPagerError(404, PAGER_ERROR.venueNotFound);
      if (settings.customerNumbers) throw new SandboxPagerError(409, PAGER_ERROR.typeYourNumber);
      const code = freeCode();
      lastId += 1;
      const ticket: SandboxTicket = {
        id: String(lastId),
        code,
        status: 'waiting',
        createdAt: clock(),
        publicToken: `sandbox-${lastId}`,
      };
      tickets.push(ticket);
      return {
        publicToken: ticket.publicToken,
        code,
        status: ticket.status,
      };
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
        status: ticket.status,
        venueName: venueName(),
        ahead,
        readyAt: ticket.readyAt,
        remindedAt: ticket.remindedAt,
        acknowledgedAt: ticket.acknowledgedAt,
        expectedReadyAt: ticket.expectedReadyAt,
        reviewUrl: settings.reviewUrl || undefined,
        ringtone: settings.ringtone,
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

    labelTicket: async ({ id, label: raw, prepMinutes }) => {
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
      ticket.expectedReadyAt = expectedAt(prepMinutes ?? settings.prepMinutes);
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

    updateSettings: async ({
      enabled,
      labelMode,
      reviewUrl,
      prepMinutes,
      customerNumbers,
      ringtone,
    }) => {
      if (reviewUrl !== undefined) {
        const link = reviewUrl.trim();
        if (link && !/^https:\/\/[^\s/@]+/.test(link)) {
          throw new SandboxPagerError(400, PAGER_ERROR.invalidReviewUrl);
        }
        settings.reviewUrl = link;
      }
      if (enabled !== undefined) settings.enabled = enabled;
      if (labelMode !== undefined) settings.labelMode = labelMode;
      if (prepMinutes !== undefined) settings.prepMinutes = prepMinutes;
      if (customerNumbers !== undefined) settings.customerNumbers = customerNumbers;
      if (ringtone !== undefined) settings.ringtone = ringtone;
      return { ...settings };
    },
  };
}
