import type { CaretMessage } from "./message";

export interface CaretChat {
  id: string;
  title: string;
  summary?: string;
  createdAt: string;
  updatedAt: string;
  messages: CaretMessage[];
  pinned?: boolean;
  archived?: boolean;
  projectId?: string;
}

export interface CaretChatSummary {
  id: string;
  title: string;
  summary?: string;
  updatedAt: string;
  pinned?: boolean;
  archived?: boolean;
  projectId?: string;
}

export interface CaretProject {
  id: string;
  name: string;
  icon: string;
  color: string;
  context: string;
  updatedAt: string;
}
