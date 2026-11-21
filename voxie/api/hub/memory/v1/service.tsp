import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.memory.v1;

@route("/hub/memory/v1")
interface MemoryService {
  @get
  @route("/chat-projects")
  listProjects(...ListProjectsRequest): ListProjectsResponse;

  @post
  @route("/chat-projects")
  saveProject(...SaveProjectRequest): SaveProjectResponse;

  @delete
  @route("/chat-projects/{id}")
  archiveProject(...ArchiveProjectRequest): ArchiveProjectResponse;

  @put
  @route("/chats/{session_id}/project")
  setChatProject(...SetChatProjectRequest): SetChatProjectResponse;

  @get
  @route("/caret-memories")
  listMemories(...ListMemoriesRequest): ListMemoriesResponse;

  @post
  @route("/caret-memories")
  saveMemory(...SaveMemoryRequest): SaveMemoryResponse;

  @delete
  @route("/caret-memories/{id}")
  deleteMemory(...DeleteMemoryRequest): DeleteMemoryResponse;

  @get
  @route("/chats")
  listChats(...ListChatsRequest): ListChatsResponse;

  @get
  @route("/chats/{session_id}/messages")
  getChatMessages(...GetChatMessagesRequest): GetChatMessagesResponse;

  @delete
  @route("/chats/{session_id}")
  deleteChat(...DeleteChatRequest): DeleteChatResponse;

  @put
  @route("/chats/{session_id}")
  renameChat(...RenameChatRequest): RenameChatResponse;

  @put
  @route("/chats/{session_id}/pin")
  pinChat(...PinChatRequest): PinChatResponse;
}
