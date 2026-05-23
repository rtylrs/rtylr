/** Hard-coded Caret feature flags. Eventually backed by the platform's
 *  feature-flag service - for now this is a single source of truth. */
export const CARET_FEATURES = {
  /** Show the speech-to-text mic button. */
  voiceInput: true,
  /** Show the file attachment button (server still needs UPLOAD endpoint). */
  attachments: false,
  /** Enable the dedicated Caret pages (history/settings/tools/overview). */
  dashboardPages: true,
  /** Show the structured thinking block when the model emits one. */
  thinkingBlocks: true,
  /** Enable tool calling. */
  toolCalls: true,
  /** Require explicit approval before executing write-scope tools. */
  approvalRequiredForWrites: true,
} as const;
