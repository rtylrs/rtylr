import { describe, expect, it } from 'vitest';
import {
  CARET_SLASH_COMMANDS,
  filterCaretSlashCommands,
  getCaretSlashCommand,
  getCaretSlashCommandQuery,
} from './slashCommands';

describe('Caret slash commands', () => {
  it('offers matching commands while typing a slash command', () => {
    expect(getCaretSlashCommandQuery('/go')).toBe('go');
    expect(filterCaretSlashCommands('/go').map((command) => command.name)).toEqual(['goal']);
    expect(filterCaretSlashCommands('/')).toHaveLength(CARET_SLASH_COMMANDS.length);
  });

  it('does not treat a command with arguments as an unfinished picker query', () => {
    expect(getCaretSlashCommandQuery('/goal launch the new store')).toBeNull();
    expect(getCaretSlashCommand('/goal launch the new store')?.mode).toBe('ask');
  });

  it('resolves commands case-insensitively', () => {
    expect(getCaretSlashCommand('  /BRIEF')?.name).toBe('brief');
    expect(getCaretSlashCommand('summarize this')).toBeNull();
  });
});
