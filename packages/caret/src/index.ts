// Panel + button — primary consumer entrypoints
export { default as CaretPanel } from './components/CaretPanel';
export type { CaretPanelProps } from './components/CaretPanel';
export { default as CaretButton } from './components/CaretButton';
export { default as CaretMessage } from './components/CaretMessage';

// Composition pieces (advanced consumers can build their own layouts)
export { default as HeaderBar } from './components/header/HeaderBar';
export { default as MessageList } from './components/MessageList';
export { default as MessageBubble } from './components/MessageBubble';
export { default as MessageBlocks } from './components/MessageBlocks';
export { default as MessageActions } from './components/MessageActions';
export { default as EmptyState } from './components/EmptyState';
export { default as SettingsPanel } from './components/SettingsPanel';
export { default as ApprovalDialog } from './components/ApprovalDialog';

export * from './components/blocks';
export * from './components/composer';
export * from './components/sidebar';

// Stores
export * from './stores';

// Types
export * from './types';

// Hooks
export * from './hooks';

// Protocol
export * from './protocol';

// Markdown (escape-hatch for embedding renderer elsewhere)
export { renderMarkdown, renderInlineMarkdown } from './markdown';
