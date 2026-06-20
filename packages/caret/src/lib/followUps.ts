import type { CaretPageContext } from '../types/context';
import type { CaretMode } from '../types/mode';
import type { CaretMessage, ContentBlock } from '../types/message';
import type { CaretStarter } from './suggestions';

interface BuildCaretFollowUpsInput {
  appId: string;
  pageContext?: CaretPageContext;
  messages: CaretMessage[];
}

const starter = (prompt: string, mode: CaretMode = 'ask'): CaretStarter => ({ prompt, mode });

function normalize(value: unknown): string {
  return typeof value === 'string' ? value.toLowerCase() : '';
}

function blockText(block: ContentBlock): string {
  if (block.type === 'text' || block.type === 'thinking') return block.content;
  if (block.summary) return block.summary;
  if (block.error) return block.error;
  return block.name;
}

function messageText(message: CaretMessage): string {
  return normalize(
    message.blocks.length > 0 ? message.blocks.map(blockText).join('\n') : message.content
  );
}

function hasPendingTool(message: CaretMessage): boolean {
  return message.blocks.some((block) => block.type === 'tool_call' && block.status === 'pending');
}

function unique(values: CaretStarter[], limit: number): CaretStarter[] {
  const seen = new Set<string>();
  const out: CaretStarter[] = [];
  for (const item of values) {
    const key = item.prompt.toLowerCase();
    if (seen.has(key)) continue;
    seen.add(key);
    out.push(item);
    if (out.length >= limit) break;
  }
  return out;
}

export function buildCaretFollowUps({
  appId,
  pageContext,
  messages,
}: BuildCaretFollowUpsInput): CaretStarter[] {
  const last = messages[messages.length - 1];
  if (!last || last.role !== 'assistant' || hasPendingTool(last)) return [];

  const text = messageText(last);
  if (!text.trim()) return [];

  const app = normalize(pageContext?.app ?? appId);
  const moduleName = normalize(pageContext?.module);
  const page = normalize(pageContext?.page);
  const context = `${app} ${moduleName} ${page} ${text}`;
  const items: CaretStarter[] = [];

  if (/error|couldn.t|failed|missing|unavailable|not enough/.test(context)) {
    items.push(starter('What information is missing?'), starter('Try a safer version', 'support'));
  }

  if (/revenue|sales|orders|cash|invoice|payment|expense/.test(context)) {
    items.push(
      starter('What changed since yesterday?', 'analyst'),
      starter('Find the biggest money risk', 'analyst')
    );
  }

  if (/inventory|stock|product|supplier|warehouse|reorder|menu/.test(context)) {
    items.push(
      starter('Show what needs restocking first', 'analyst'),
      starter('Turn this into a reorder plan', 'agent')
    );
  }

  if (/customer|crm|loyalty|campaign|segment/.test(context)) {
    items.push(
      starter('Which customers need attention?', 'analyst'),
      starter('Suggest the next customer action', 'agent')
    );
  }

  if (/employee|staff|team|shift|attendance|payroll|leave/.test(context)) {
    items.push(
      starter('Show team issues to handle today', 'analyst'),
      starter('What should the manager do next?', 'agent')
    );
  }

  if (/task|project|roadmap|blocked|overdue|workload/.test(context)) {
    items.push(
      starter('Find blocked work', 'analyst'),
      starter('Create a short action plan', 'agent')
    );
  }

  items.push(
    starter('What should I do next?'),
    starter('Find anything risky', 'analyst'),
    starter('Turn this into an action plan', 'agent')
  );

  return unique(items, 3);
}
