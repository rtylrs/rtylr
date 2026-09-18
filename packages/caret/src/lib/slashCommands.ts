import type { CaretMode } from '../types/mode';

export interface CaretSlashCommand {
  name: string;
  description: string;
  usage: string;
  mode: CaretMode;
}

/** Commands understood by the Caret backend. Keep this list small and task-oriented. */
export const CARET_SLASH_COMMANDS: CaretSlashCommand[] = [
  {
    name: 'goal',
    description: 'Turn an outcome into a clear goal and next steps.',
    usage: '/goal <outcome>',
    mode: 'ask',
  },
  {
    name: 'brief',
    description: 'Give me a concise owner brief for the current business.',
    usage: '/brief',
    mode: 'analyst',
  },
  {
    name: 'status',
    description: 'Summarize what is happening and what needs attention.',
    usage: '/status',
    mode: 'analyst',
  },
  {
    name: 'remember',
    description: 'Save a business preference or operating rule.',
    usage: '/remember <fact>',
    mode: 'agent',
  },
  {
    name: 'help',
    description: 'Show what Caret can do and how to use these commands.',
    usage: '/help',
    mode: 'support',
  },
];

/** Return the slash-command token while the composer is still choosing a command. */
export function getCaretSlashCommandQuery(value: string): string | null {
  const match = /^\s*\/([^\s]*)$/.exec(value);
  return match ? match[1].toLowerCase() : null;
}

export function filterCaretSlashCommands(value: string): CaretSlashCommand[] {
  const query = getCaretSlashCommandQuery(value);
  if (query === null) return [];

  return CARET_SLASH_COMMANDS.filter((command) => command.name.startsWith(query));
}

/** Resolve the command at the beginning of a submitted prompt, if any. */
export function getCaretSlashCommand(prompt: string): CaretSlashCommand | null {
  const match = /^\s*\/([^\s]+)/.exec(prompt);
  if (!match) return null;

  const name = match[1].toLowerCase();
  return CARET_SLASH_COMMANDS.find((command) => command.name === name) ?? null;
}
