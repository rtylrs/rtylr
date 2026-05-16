/** Known slash-commands for the Caret composer. */
export interface SlashCommandDef {
  command: string;
  description: string;
  example: string;
}

export const SLASH_COMMANDS: SlashCommandDef[] = [
  { command: 'help', description: 'List available shortcuts and commands.', example: '/help' },
  { command: 'clear', description: 'Start a fresh chat.', example: '/clear' },
  {
    command: 'mode',
    description: 'Switch Caret mode (ask/agent/analyst).',
    example: '/mode agent',
  },
  { command: 'model', description: 'Switch model provider.', example: '/model gpt-4o' },
  { command: 'tools', description: 'Show available tools.', example: '/tools' },
  { command: 'export', description: 'Download the current chat transcript.', example: '/export' },
];
