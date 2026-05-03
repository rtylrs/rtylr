#!/usr/bin/env node
import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from '@modelcontextprotocol/sdk/types.js';

import { flowTools } from './tools/flow.js';
import { crmTools } from './tools/crm.js';
import { financeTools } from './tools/finance.js';
import { erpTools } from './tools/erp.js';

const ALL_TOOLS = [...flowTools, ...crmTools, ...financeTools, ...erpTools];

const server = new Server(
  { name: 'rtylr-mcp', version: '0.1.0' },
  { capabilities: { tools: {} } }
);

server.setRequestHandler(ListToolsRequestSchema, async () => ({
  tools: ALL_TOOLS.map(({ handler: _h, ...tool }) => tool),
}));

server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const tool = ALL_TOOLS.find((t) => t.name === request.params.name);

  if (!tool) {
    return {
      content: [{ type: 'text', text: `Unknown tool: ${request.params.name}` }],
      isError: true,
    };
  }

  const input = (request.params.arguments ?? {}) as Record<string, unknown>;
  return tool.handler(input);
});

async function main(): Promise<void> {
  const transport = new StdioServerTransport();
  await server.connect(transport);
}

main().catch((err) => {
  process.stderr.write(`Fatal: ${err instanceof Error ? err.message : String(err)}\n`);
  process.exit(1);
});
