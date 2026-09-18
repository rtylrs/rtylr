import type { CaretMessage } from './message';

export interface CaretChat {
  id: string;
  title: string;
  summary?: string;
  createdAt: string;
  updatedAt: string;
  messages: CaretMessage[];
  pinned?: boolean;
  archived?: boolean;
}

export interface CaretChatSummary {
  id: string;
  title: string;
  summary?: string;
  updatedAt: string;
  pinned?: boolean;
  archived?: boolean;
}
