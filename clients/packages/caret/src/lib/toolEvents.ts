import type { ToolCallBlock } from "../types/message";

export interface CaretToolSuccess {
  name: string;
  input: Record<string, unknown>;
  output: unknown;
  /** The tool changed something the person approved, such as a record. */
  approved: boolean;
}

type Listener = (event: CaretToolSuccess) => void;

const listeners = new Set<Listener>();

/** Calls `listener` each time a Caret tool finishes successfully, in any
 *  chat, so the host app can refresh what the tool changed. Returns the
 *  unsubscribe function. */
export function onCaretToolSuccess(listener: Listener): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export function emitCaretToolSuccess(block: ToolCallBlock, output: unknown): void {
  const event = {
    name: block.name,
    input: block.input ?? {},
    output,
    approved: Boolean(block.requiresApproval),
  };
  for (const listener of listeners) {
    try {
      listener(event);
    } catch {
      // A broken listener must not break the chat.
    }
  }
}
