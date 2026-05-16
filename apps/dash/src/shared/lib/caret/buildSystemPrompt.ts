import { CARET_MODES, type CaretMode } from '@voxire/caret';

/** Assemble the full system prompt the backend should use. Today the
 *  backend constructs this itself — this helper exists so the dash side
 *  can surface a preview/debug view of the actual prompt that will be sent. */
export interface BuildSystemPromptInput {
  mode: CaretMode;
  systemContext: string;
}

export function buildSystemPromptPreview({ mode, systemContext }: BuildSystemPromptInput): string {
  const def = CARET_MODES[mode];
  return [
    'You are Caret, an AI business assistant for rtylr.',
    `Today is ${new Date().toLocaleDateString()}.`,
    `Current context: ${systemContext}`,
    def.systemPromptAddendum,
  ].join('\n\n');
}
