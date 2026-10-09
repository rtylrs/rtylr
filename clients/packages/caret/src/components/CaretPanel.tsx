import React, { useCallback, useEffect, useMemo, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import "../styles/index.css";
import { caretTokens } from "../ui/tokens";
import { useCaretStore } from "../stores/caretStore";
import { useCaretSettingsStore } from "../stores/settingsStore";
import { useCaretUIStore } from "../stores/uiStore";
import { useChatsStore } from "../stores/chatsStore";
import { useCaretStream } from "../hooks/useCaretStream";
import { useCaretSessions } from "../hooks/useCaretSessions";
import { useCaretProjects } from "../hooks/useCaretProjects";
import { useLoadCaretChat } from "../hooks/useLoadCaretChat";
import { usePanelResize } from "../hooks/usePanelResize";
import { useShortcuts } from "../hooks/useShortcuts";
import { authHeaders } from "../lib/auth";
import { createMemoryClient } from "../lib/memoryClient";
import { ApiError } from "@voxire/voxie/hub/memory/v1/client";
import { buildCaretEmptyState, type CaretStarter } from "../lib/suggestions";
import type { CaretChatSummary, CaretProject } from "../types/chat";
import type { CaretPageContext } from "../types/context";
import type { CaretMessage, ToolCallBlock } from "../types/message";
import HeaderBar from "./header/HeaderBar";
import MessageList from "./MessageList";
import MessageSkeleton from "./MessageSkeleton";
import { Composer } from "./composer";
import WelcomeHero from "./WelcomeHero";
import SettingsPanel from "./SettingsPanel";
import { ChatList, type ChatListLabels } from "./sidebar";
import ProjectSidebar, { type ProjectLabels } from "./sidebar/ProjectSidebar";

export interface CaretWorkspaceOptions {
  header: React.ReactNode;
  sidebarHeader: React.ReactNode;
  welcome: React.ReactNode;
  historyLabels: ChatListLabels;
  projectLabels?: Partial<ProjectLabels>;
  composerPlaceholder: string;
  sidebarOpen: boolean;
  onSelectChat: () => void;
}

export interface CaretPanelProps {
  appId: string;
  workspace?: CaretWorkspaceOptions;
  shortcutsEnabled?: boolean;
  systemContext?: string;
  pageContext?: CaretPageContext;
  userId: number;
  apiBaseUrl: string;
  /**
   * 'inline' renders inside a flex column the parent controls (a side panel).
   * 'fullscreen' fills its container edge-to-edge as a stand-alone page/route.
   * 'overlay' floats as a centered, rounded panel over a dimmed backdrop -
   * summoned from an "Ask Caret" trigger elsewhere on the page.
   * Both 'fullscreen' and 'overlay' show a centered welcome hero when empty
   * and a centered chat column once messages exist.
   */
  variant?: "inline" | "fullscreen" | "overlay";
}

type ToolDecision = "approve" | "deny";

const EMPTY_MESSAGES: CaretMessage[] = [];

interface ToolDecisionResponse {
  tool_call_id: string;
  status: string;
  message: string;
  executed: boolean;
  preview?: string;
  output?: unknown;
}

function normalizeToolDecisionStatus(status: string): ToolCallBlock["status"] {
  switch (status) {
    case "pending":
    case "approved":
    case "denied":
    case "running":
    case "success":
    case "error":
      return status;
    default:
      return "error";
  }
}

function outputMessage(output: unknown): string | null {
  if (!output || typeof output !== "object") return null;
  const message = (output as { message?: unknown }).message;
  return typeof message === "string" && message.trim() ? message.trim() : null;
}

function cleanDecisionMessage(message: string | null | undefined): string {
  const text = (message ?? "").trim();
  if (!text || text === "Tool executed successfully.")
    return "the action is complete.";
  return text.replace(/\.$/, "") + ".";
}

function decisionFollowUp(
  decision: ToolDecision,
  status: ToolCallBlock["status"],
  result?: ToolDecisionResponse,
): string | null {
  if (decision === "deny" || status === "denied") {
    return "Okay, I didn't make any changes.";
  }
  if (status === "success" && result?.executed) {
    return `Done. ${cleanDecisionMessage(outputMessage(result.output) ?? result.message)}`;
  }
  if (status === "error") {
    return "I couldn't finish that. No confirmed changes were made.";
  }
  return null;
}

const CaretPanel: React.FC<CaretPanelProps> = ({
  appId,
  systemContext,
  pageContext,
  userId,
  apiBaseUrl,
  variant,
  workspace,
  shortcutsEnabled = true,
}) => {
  useShortcuts(shortcutsEnabled);
  const caret = useCaretStore();
  const activeTab =
    caret.tabs.find((tab) => tab.id === caret.activeTabId) ?? caret.tabs[0];
  const activeTabId = activeTab?.id ?? caret.activeTabId;
  const sessionId = activeTab?.sessionId ?? null;
  const messages = activeTab?.messages ?? EMPTY_MESSAGES;
  const isStreaming = activeTab?.isStreaming ?? false;
  const isLoading = activeTab?.isLoading ?? false;
  const { isOpen } = caret;
  const showThinking = useCaretSettingsStore((s) => s.showThinking);
  const panelWidth = useCaretSettingsStore((s) => s.panelWidth);
  const view = useCaretUIStore((s) => s.view);
  const setView = useCaretUIStore((s) => s.setView);
  const queryClient = useQueryClient();
  const mode = "agent" as const;
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(
    activeTab?.projectId ?? null,
  );

  const { sendMessage, stopStreaming } = useCaretStream({
    userId,
    apiBaseUrl,
    systemContext,
    pageContext,
  });
  const {
    sessions,
    isLoading: sessionsLoading,
    error: sessionsError,
    refetch: refetchSessions,
  } = useCaretSessions({ userId, apiBaseUrl });
  const {
    projects,
    isLoading: projectsLoading,
    error: projectsError,
    refetch: refetchProjects,
  } = useCaretProjects({ userId, apiBaseUrl });
  const retrySessions = useCallback((): void => {
    void refetchSessions();
  }, [refetchSessions]);
  const loadChat = useLoadCaretChat({ userId, apiBaseUrl });
  const setRecent = useChatsStore((s) => s.setRecent);
  const removeRecent = useChatsStore((s) => s.removeRecent);
  const { onMouseDown } = usePanelResize();
  const emptyState = useMemo(
    () => buildCaretEmptyState({ appId, pageContext }),
    [appId, pageContext],
  );

  useEffect(() => {
    if (sessions.length > 0) setRecent(sessions);
  }, [sessions, setRecent]);

  useEffect(() => {
    setSelectedProjectId(activeTab?.projectId ?? null);
  }, [activeTab?.projectId]);

  const createChat = useCallback(
    (projectId: string | null): void => {
      caret.newTab({ projectId });
      setView("chat");
      workspace?.onSelectChat();
    },
    [caret, setView, workspace],
  );
  const newChat = useCallback((): void => {
    createChat(selectedProjectId);
  }, [createChat, selectedProjectId]);

  const selectProject = useCallback(
    (projectId: string | null): void => {
      setSelectedProjectId(projectId);
      const current = useCaretStore
        .getState()
        .tabs.find((tab) => tab.id === useCaretStore.getState().activeTabId);
      if (current?.projectId !== projectId) {
        const existing = useCaretStore
          .getState()
          .tabs.find((tab) => tab.projectId === projectId);
        if (existing) caret.setActiveTab(existing.id);
        else createChat(projectId);
      }
      setView("chat");
    },
    [caret, createChat, setView],
  );

  const saveProject = useCallback(
    async (
      input: import("@voxire/voxie/hub/memory/v1/types").SaveProjectRequest,
    ): Promise<CaretProject> => {
      const response = await createMemoryClient(apiBaseUrl).saveProject(input);
      await queryClient.invalidateQueries({
        queryKey: ["caret", "projects", userId],
      });
      if (!response.project?.id || !response.project.name)
        throw new Error("The server returned an invalid project.");
      return {
        id: response.project.id,
        name: response.project.name,
        icon: response.project.icon || "folder",
        color: response.project.color || "neutral",
        context: response.project.context || "",
        updatedAt: response.project.updatedAt || new Date().toISOString(),
      };
    },
    [apiBaseUrl, queryClient, userId],
  );

  const archiveProject = useCallback(
    async (projectId: string): Promise<void> => {
      await createMemoryClient(apiBaseUrl).archiveProject({ id: projectId });
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ["caret", "projects", userId],
        }),
        queryClient.invalidateQueries({
          queryKey: ["caret", "sessions", userId],
        }),
      ]);
    },
    [apiBaseUrl, queryClient, userId],
  );

  const selectChat = useCallback(
    (id: string): void => {
      const summary = sessions.find((chat) => chat.id === id);
      useCaretStore.getState().openChat({
        sessionId: id,
        title: summary?.title,
        projectId: summary?.projectId,
      });
      setSelectedProjectId(
        sessions.find((chat) => chat.id === id)?.projectId ?? null,
      );
      setView("chat");
      workspace?.onSelectChat();
    },
    [caret, sessions, setView, workspace],
  );

  useEffect(() => {
    if (
      !activeTab ||
      activeTab.isLoaded ||
      activeTab.isLoading ||
      !activeTab.sessionId
    )
      return;

    const tabId = activeTab.id;
    const tabSessionId = activeTab.sessionId;
    // A prompt can be sent while the history is still in flight. Never let the
    // loaded transcript replace those messages, but do mark the tab loaded so
    // the effect doesn't retry the fetch forever.
    const applyHistory = (loaded: CaretMessage[]): void => {
      const tab = useCaretStore
        .getState()
        .tabs.find((candidate) => candidate.id === tabId);
      if (
        tab &&
        tab.sessionId === tabSessionId &&
        !tab.isStreaming &&
        tab.messages.length === 0
      ) {
        caret.setMessages(loaded, tabId);
      } else {
        caret.setTabLoaded(true, tabId);
      }
    };

    caret.setTabLoading(true, tabId);
    void loadChat(tabSessionId)
      .then(applyHistory)
      .catch(() => applyHistory([]))
      .finally(() => caret.setTabLoading(false, tabId));
  }, [activeTab, caret, loadChat]);

  const closeChatTab = useCallback(
    (tabId: string): void => {
      const tab = useCaretStore
        .getState()
        .tabs.find((candidate) => candidate.id === tabId);
      if (tab?.isStreaming) stopStreaming(tabId);
      caret.closeTab(tabId);
    },
    [caret, stopStreaming],
  );

  const deleteChat = useCallback(
    async (id: string): Promise<void> => {
      // Prefix match: the sidebar and the command palette load different limits.
      const queryKey = ["caret", "sessions", userId] as const;
      const previousLists = queryClient.getQueriesData<CaretChatSummary[]>({ queryKey });
      const previous = previousLists[0]?.[1];

      removeRecent(id);
      queryClient.setQueriesData<CaretChatSummary[]>({ queryKey }, (old) =>
        old ? old.filter((chat) => chat.id !== id) : [],
      );

      const deleted = await createMemoryClient(apiBaseUrl)
        .deleteChat({ sessionId: id })
        .then(() => true)
        .catch((err) => {
          // A 404 means the chat is already gone server-side - treat it as
          // a successful delete rather than reverting the optimistic update.
          return err instanceof ApiError && err.statusCode === 404;
        });

      if (!deleted) {
        for (const [key, list] of previousLists) queryClient.setQueryData(key, list);
        if (previous) setRecent(previous);
        return;
      }

      const openTab = useCaretStore
        .getState()
        .tabs.find((tab) => tab.sessionId === id);
      const wasActive = openTab?.id === activeTabId;
      if (openTab) {
        closeChatTab(openTab.id);
      }
      if (wasActive) {
        setView("chat");
      }

      void queryClient.invalidateQueries({ queryKey });
    },
    [
      activeTabId,
      apiBaseUrl,
      closeChatTab,
      queryClient,
      removeRecent,
      setRecent,
      setView,
      userId,
    ],
  );

  const [introPlaying, setIntroPlaying] = useState(false);
  useEffect(() => {
    if (!isOpen) return;
    try {
      if (localStorage.getItem("caret:intro-shown") === "1") return;
      localStorage.setItem("caret:intro-shown", "1");
    } catch {
      return;
    }
    setIntroPlaying(true);
    const timer = window.setTimeout(() => setIntroPlaying(false), 1500);
    return () => window.clearTimeout(timer);
  }, [isOpen]);

  const findToolCall = useCallback(
    (id: string, tabId: string): ToolCallBlock | null => {
      const tab = useCaretStore
        .getState()
        .tabs.find((candidate) => candidate.id === tabId);
      for (const message of tab?.messages ?? []) {
        for (const block of message.blocks) {
          if (block.type === "tool_call" && block.id === id) return block;
        }
      }
      return null;
    },
    [],
  );

  const sendToolDecision = useCallback(
    async (id: string, decision: ToolDecision): Promise<void> => {
      const tabId = activeTabId;
      const call = findToolCall(id, tabId);
      if (!call) {
        caret.setToolStatus(
          id,
          "error",
          { error: "Tool call no longer exists." },
          tabId,
        );
        return;
      }

      const baseUrl = apiBaseUrl.endsWith("/") ? apiBaseUrl : `${apiBaseUrl}/`;
      caret.setToolStatus(id, "running", undefined, tabId);

      try {
        const response = await fetch(`${baseUrl}caret/tools/decision`, {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            ...authHeaders(),
          },
          body: JSON.stringify({
            tool_call_id: id,
            name: call.name,
            input: call.input ?? {},
            decision,
            mode: call.mode ?? mode,
          }),
        });

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const result = (await response.json()) as ToolDecisionResponse;
        const status = normalizeToolDecisionStatus(result.status);
        const extras: Partial<ToolCallBlock> =
          status === "error"
            ? {
                error: result.message || "Tool decision failed.",
                preview: result.preview,
              }
            : {
                output: result.output ?? {
                  message: result.message,
                  executed: result.executed,
                },
                preview: result.preview,
              };
        caret.setToolStatus(id, status, extras, tabId);
        const followUp = decisionFollowUp(decision, status, result);
        if (followUp) {
          caret.appendBlockToLastAssistant(
            {
              id: `tool-followup-${id}-${Date.now()}`,
              type: "text",
              content: followUp,
            },
            tabId,
          );
        }
      } catch (error) {
        const message =
          error instanceof Error ? error.message : "Tool decision failed.";
        caret.setToolStatus(id, "error", { error: message }, tabId);
        caret.appendBlockToLastAssistant(
          {
            id: `tool-followup-${id}-${Date.now()}`,
            type: "text",
            content: "I couldn't finish that. No confirmed changes were made.",
          },
          tabId,
        );
      }
    },
    [activeTabId, apiBaseUrl, caret, findToolCall, mode],
  );

  const onApprove = useCallback(
    (id: string): void => {
      void sendToolDecision(id, "approve");
    },
    [sendToolDecision],
  );
  const onDeny = useCallback(
    (id: string): void => {
      void sendToolDecision(id, "deny");
    },
    [sendToolDecision],
  );

  const sendPromptWithIntent = useCallback(
    (prompt: string): void => {
      void sendMessage(prompt, "agent");
    },
    [sendMessage],
  );

  const sendStarter = useCallback(
    (starter: CaretStarter): void => {
      sendPromptWithIntent(starter.prompt);
    },
    [sendPromptWithIntent],
  );

  const isFullscreen = variant === "fullscreen";
  const isOverlay = variant === "overlay";
  const [expanded, setExpanded] = useState(false);
  /** Detached from the docked/compact presentation into a full, floating workspace. */
  const isDetached = !isFullscreen && expanded;
  const isBig = isFullscreen || isDetached;
  const isEmpty = !isLoading && messages.length === 0;
  const visibleSessions =
    selectedProjectId === null
      ? sessions
      : sessions.filter((chat) => chat.projectId === selectedProjectId);
  const mainShowsChat = view === "chat" || (isBig && view === "history");

  const rootStyle: React.CSSProperties = isDetached
    ? {
        position: "fixed",
        inset: "clamp(16px, 3vh, 48px) clamp(16px, 3vw, 64px)",
        zIndex: 9999,
        background: caretTokens.bg,
        boxSizing: "border-box",
        borderRadius: 20,
        border: `1px solid ${caretTokens.border}`,
        boxShadow: caretTokens.shadow,
        display: "flex",
        flexDirection: "column",
        overflow: "hidden",
        transition: "inset 0.18s ease",
      }
    : variant === "inline" || isFullscreen
      ? {
          position: "relative",
          width: "100%",
          height: "100%",
          background: caretTokens.bg,
          boxSizing: "border-box",
          minWidth: 0,
          borderInlineStart: isFullscreen
            ? "none"
            : `1px solid ${caretTokens.border}`,
          display: "flex",
          flexDirection: "column",
          overflow: "hidden",
        }
      : isOverlay
        ? {
            position: "relative",
            width: "min(880px, 92vw)",
            height: "min(760px, 86vh)",
            background: caretTokens.bg,
            boxSizing: "border-box",
            minWidth: 0,
            borderRadius: 20,
            border: `1px solid ${caretTokens.border}`,
            boxShadow: caretTokens.shadow,
            display: "flex",
            flexDirection: "column",
            overflow: "hidden",
            transform: isOpen
              ? "scale(1) translateY(0)"
              : "scale(0.97) translateY(8px)",
            transition: "transform 0.18s ease",
          }
        : {
            position: "fixed",
            top: 0,
            right: 0,
            width: panelWidth,
            height: "100dvh",
            zIndex: 9999,
            background: caretTokens.bg,
            boxSizing: "border-box",
            minWidth: 0,
            borderLeft: `1px solid ${caretTokens.border}`,
            display: "flex",
            flexDirection: "column",
            transform: isOpen ? "translateX(0)" : "translateX(100%)",
            transition: "transform 0.25s cubic-bezier(0.4, 0, 0.2, 1)",
            willChange: "transform",
          };

  const centerColStyle: React.CSSProperties | undefined = isBig
    ? {
        width: "100%",
        maxWidth: 820,
        margin: "0 auto",
        padding: "0 24px",
        boxSizing: "border-box",
      }
    : undefined;
  const composerWrapStyle: React.CSSProperties | undefined = isBig
    ? { ...centerColStyle, paddingBottom: 20 }
    : undefined;

  const mainColumn = (
    <div
      className="caret-main-column"
      style={{
        display: "flex",
        flexDirection: "column",
        flex: 1,
        minHeight: 0,
        overflow: "hidden",
      }}
    >
      {workspace?.header ?? (
        <HeaderBar
          onNewChat={newChat}
          hideClose={isFullscreen}
          title={activeTab?.title}
          expanded={!isFullscreen ? expanded : undefined}
          onToggleExpand={
            !isFullscreen ? () => setExpanded((v) => !v) : undefined
          }
        />
      )}
      {mainShowsChat && (
        <>
          <div
            id={`caret-chat-panel-${activeTabId}`}
            className="caret-conversation"
            style={{
              display: "flex",
              flex: 1,
              flexDirection: "column",
              minHeight: 0,
              overflow: "hidden",
              justifyContent: workspace && isEmpty ? "center" : undefined,
            }}
          >
            <div
              style={{
                flex: workspace && isEmpty ? undefined : 1,
                minHeight: 0,
                display: "flex",
                flexDirection: "column",
                ...centerColStyle,
              }}
            >
              {isEmpty ? (
                (workspace?.welcome ?? (
                  <WelcomeHero
                    size={isBig ? "lg" : "sm"}
                    emptyState={emptyState}
                    onStarter={sendStarter}
                  />
                ))
              ) : isLoading ? (
                <div style={{ flex: 1, overflowY: "auto", padding: 18 }}>
                  <MessageSkeleton />
                </div>
              ) : (
                <MessageList
                  messages={messages}
                  isStreaming={isStreaming}
                  scrollKey={activeTabId}
                  showThinking={showThinking}
                  onApproveTool={onApprove}
                  onDenyTool={onDeny}
                />
              )}
            </div>
            <div
              className="caret-composer-wrap"
              style={{
                ...composerWrapStyle,
                ...(workspace && isEmpty ? { paddingBottom: 0 } : {}),
              }}
            >
              <Composer
                key={activeTabId}
                autoFocus={isOpen || isFullscreen}
                size={isBig ? "lg" : "md"}
                isStreaming={isStreaming}
                onSubmit={(t, images) => {
                  void sendMessage(t, "agent", images);
                }}
                onStop={() => stopStreaming(activeTabId)}
                placeholder={
                  workspace?.composerPlaceholder ??
                  emptyState.composerPlaceholder
                }
              />
            </div>
          </div>
        </>
      )}

      {!isBig && view === "history" && (
        <ChatList
          chats={visibleSessions}
          isLoading={sessionsLoading}
          error={sessionsError}
          onRetry={retrySessions}
          activeId={sessionId}
          onSelect={selectChat}
          onNew={newChat}
          onDelete={deleteChat}
        />
      )}

      {view === "settings" && <SettingsPanel />}
    </div>
  );

  const panel = (
    <aside
      role="complementary"
      aria-label="Caret AI assistant"
      inert={!isOpen && !isFullscreen}
      aria-hidden={!isOpen && !isFullscreen}
      className={
        (introPlaying ? "caret-panel-shell caret-intro" : "caret-panel-shell") +
        (isBig ? " caret-panel-textured" : "")
      }
      style={rootStyle}
    >
      {variant === "inline" && !isDetached && (
        <span
          role="separator"
          aria-orientation="vertical"
          aria-label="Resize Caret panel"
          onMouseDown={onMouseDown}
          style={{
            position: "absolute",
            top: 0,
            left: -3,
            width: 6,
            height: "100%",
            cursor: "col-resize",
            zIndex: 10,
          }}
        />
      )}

      {isBig ? (
        <div
          style={{ display: "flex", flex: 1, minHeight: 0, overflow: "hidden" }}
        >
          <div
            className="caret-history-sidebar"
            id={workspace ? "caret-chat-history" : undefined}
            hidden={workspace && !workspace.sidebarOpen}
            style={{
              width: 260,
              flexShrink: 0,
              minHeight: 0,
              borderInlineEnd: `1px solid ${caretTokens.border}`,
              display: workspace && !workspace.sidebarOpen ? "none" : "flex",
              flexDirection: "column",
              overflow: "hidden",
              background: caretTokens.bg,
            }}
          >
            {workspace?.sidebarHeader}
            <ProjectSidebar
              projects={projects}
              activeId={selectedProjectId}
              labels={workspace?.projectLabels}
              isLoading={projectsLoading}
              loadError={projectsError !== null}
              onSelect={selectProject}
              onSave={saveProject}
              onArchive={archiveProject}
            />
            <ChatList
              labels={workspace?.historyLabels}
              chats={visibleSessions}
              isLoading={sessionsLoading}
              error={sessionsError}
              onRetry={retrySessions}
              activeId={sessionId}
              onSelect={selectChat}
              onNew={newChat}
              onDelete={deleteChat}
            />
          </div>
          {mainColumn}
        </div>
      ) : (
        mainColumn
      )}
    </aside>
  );

  if (!isOverlay && !isDetached) return panel;

  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        zIndex: 9999,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: isOpen ? "rgba(0, 0, 0, 0.45)" : "transparent",
        backdropFilter: isOpen ? "blur(6px)" : "none",
        WebkitBackdropFilter: isOpen ? "blur(6px)" : "none",
        opacity: isOpen ? 1 : 0,
        pointerEvents: isOpen ? "auto" : "none",
        transition:
          "opacity 0.18s ease, background 0.18s ease, backdrop-filter 0.18s ease",
      }}
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) caret.close();
      }}
    >
      {panel}
    </div>
  );
};

export default CaretPanel;
