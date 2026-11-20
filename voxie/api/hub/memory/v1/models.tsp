import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.memory.v1;

model ChatSummary {
  id: string;
  title: string;
  summary: string;
  updated_at: string;
  pinned: boolean;
  archived: boolean;
  project_id: string;
}

model ChatMessage {
  id: string;
  role: string;
  content: string;
  created_at: string;
  tokens_used: int32;
  attachments_json: string;
  blocks_json: string;
  response_ms: int32;
}

model ListChatsRequest {
  @query("limit")
  limit: int32;

  @query("offset")
  offset: int32;
}

model ListChatsResponse {
  chats: ChatSummary[];
}

model GetChatMessagesRequest {
  session_id: string;
}

model GetChatMessagesResponse {
  messages: ChatMessage[];
}

model DeleteChatRequest {
  session_id: string;
}

model DeleteChatResponse {
  deleted: boolean;
}

model RenameChatRequest {
  session_id: string;
  title: string;
}

model RenameChatResponse {
  updated: boolean;
}

model PinChatRequest {
  session_id: string;
  pinned: boolean;
}

model PinChatResponse {
  updated: boolean;
}

model CaretProject {
  id: string;
  name: string;
  icon: string;
  color: string;
  context: string;
  updated_at: string;
}

model ListProjectsRequest {}

model ListProjectsResponse {
  projects: CaretProject[];
}

model SaveProjectRequest {
  id: string;
  name: string;
  icon: string;
  color: string;
  context: string;
}

model SaveProjectResponse {
  project: CaretProject;
}

model ArchiveProjectRequest {
  id: string;
}

model ArchiveProjectResponse {
  archived: boolean;
}

model SetChatProjectRequest {
  session_id: string;
  project_id: string;
}

model SetChatProjectResponse {
  updated: boolean;
}

model CaretMemory {
  id: string;
  type: string;
  value: string;
  created_at: string;
  updated_at: string;

  /**
   * How Caret came to know it: "feedback" (distilled from rated replies),
   * "chat" (picked up from a conversation) or empty (told or approved).
   */
  origin: string;
}

model ListMemoriesRequest {}

model ListMemoriesResponse {
  memories: CaretMemory[];
}

model SaveMemoryRequest {
  id: string;
  value: string;
}

model SaveMemoryResponse {
  memory: CaretMemory;
}

model DeleteMemoryRequest {
  id: string;
}

model DeleteMemoryResponse {
  deleted: boolean;
}
