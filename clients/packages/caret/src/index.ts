// Panel + button - primary consumer entrypoints
export { default as CaretPanel } from "./components/CaretPanel";
export type { CaretPanelProps } from "./components/CaretPanel";
export { default as CaretButton } from "./components/CaretButton";
export { default as CaretIcon } from "./components/CaretIcon";
export { default as CaretLogo } from "./components/CaretLogo";
export { default as CaretMessage } from "./components/CaretMessage";

// Composition pieces (advanced consumers can build their own layouts)
export { default as HeaderBar } from "./components/header/HeaderBar";
export { default as ChatTabs } from "./components/ChatTabs";
export type { ChatTabsProps } from "./components/ChatTabs";
export { default as MessageList } from "./components/MessageList";
export { default as MessageBubble } from "./components/MessageBubble";
export { default as MessageBlocks } from "./components/MessageBlocks";
export { default as MessageActions } from "./components/MessageActions";
export { default as WelcomeHero } from "./components/WelcomeHero";
export type { WelcomeHeroProps } from "./components/WelcomeHero";
export { default as IntroCards } from "./components/IntroCards";
export type { IntroCardsProps } from "./components/IntroCards";
export { default as SettingsPanel } from "./components/SettingsPanel";
export { default as ApprovalDialog } from "./components/ApprovalDialog";
export { default as CaretOrb, orbStateForTool } from "./components/CaretOrb";
export type { CaretOrbProps, CaretOrbState } from "./components/CaretOrb";

export {
  Markdown,
  CodeBlock,
  ToolCallCard,
  StreamingDot,
  FormMappingPlan,
} from "./components/blocks";
export type {
  MarkdownProps,
  CodeBlockProps,
  ToolCallCardProps,
  FormMappingPlanProps,
} from "./components/blocks";
// Note: ThinkingBlock is re-exported only as a value (the message block type
// of the same name lives in ./types).
export { ThinkingBlock as ThinkingBlockView } from "./components/blocks";
export type { ThinkingBlockProps } from "./components/blocks";

export * from "./components/composer";
export * from "./components/sidebar";

// Stores
export * from "./stores";

// Types - re-export with friendlier aliases for legacy consumers.
export type {
  MessageRole,
  CaretMessage as CaretMessageType,
  ContentBlock,
  TextBlock,
  ThinkingBlock,
  ToolCallBlock,
  CaretChat,
  CaretChatSummary,
  CaretMode,
  CaretModeDef,
  CaretModel,
  CaretModelDef,
} from "./types";
export type { CaretPageContext, CaretPageContextValue } from "./types/context";
export {
  CARET_MODES,
  DEFAULT_MODE,
  CARET_MODELS,
  DEFAULT_MODEL,
} from "./types";

// Hooks
export * from "./hooks";

// Protocol
export * from "./protocol";

// Host integration: the app around Caret renders page cards and reacts to
// what Caret's tools change.
export { onCaretToolSuccess } from "./lib/toolEvents";
export type { CaretToolSuccess } from "./lib/toolEvents";
export { CaretHostProvider, useCaretHost } from "./host";
export type { CaretHost, CaretPageSpec } from "./host";

// Markdown (escape-hatch for embedding renderer elsewhere)
export { renderMarkdown, renderInlineMarkdown } from "./markdown";

// Context helpers
export { buildCaretRouteContext } from "./lib/routeContext";
export type { BuildCaretRouteContextInput } from "./lib/routeContext";
export {
  CARET_SLASH_COMMANDS,
  filterCaretSlashCommands,
  getCaretSlashCommand,
  getCaretSlashCommandQuery,
} from "./lib/slashCommands";
export type { CaretSlashCommand } from "./lib/slashCommands";
