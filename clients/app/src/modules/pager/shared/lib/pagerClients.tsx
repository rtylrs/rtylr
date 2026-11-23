import React, { createContext, useContext } from 'react';
import type { PagerServiceClient } from '@voxire/voxie/pager/v1/client';
import { publicPagerClient, staffPagerClient } from './clients';

export type PublicPagerClient = Pick<
  PagerServiceClient,
  'getPublicVenue' | 'scanTicket' | 'getPublicTicket' | 'acknowledgeTicket'
>;

export type StaffPagerClient = Pick<
  PagerServiceClient,
  | 'listTickets'
  | 'labelTicket'
  | 'markReady'
  | 'remindTicket'
  | 'undoReady'
  | 'collectTicket'
  | 'getSettings'
  | 'updateSettings'
>;

interface PagerClients {
  publicClient: PublicPagerClient;
  staffClient: StaffPagerClient;
  /** The public demo: no signed-in business behind the board. */
  sandbox: boolean;
}

const PagerClientsContext = createContext<PagerClients | null>(null);

/** Points the pager screens below at other clients, such as the demo's in-memory one. */
export function PagerClientsProvider({
  value,
  children,
}: {
  value: PagerClients;
  children: React.ReactNode;
}): React.JSX.Element {
  return <PagerClientsContext.Provider value={value}>{children}</PagerClientsContext.Provider>;
}

// Without a provider the screens talk to the real API, so the routes need no wrapper.
export function usePublicPagerClient(): PublicPagerClient {
  return useContext(PagerClientsContext)?.publicClient ?? publicPagerClient;
}

export function useStaffPagerClient(): StaffPagerClient {
  return useContext(PagerClientsContext)?.staffClient ?? staffPagerClient;
}

export function usePagerSandbox(): boolean {
  return useContext(PagerClientsContext)?.sandbox ?? false;
}
