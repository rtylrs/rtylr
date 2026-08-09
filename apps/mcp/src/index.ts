#!/usr/bin/env node
import { createServer as createHttpServer } from 'node:http';
import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { StreamableHTTPServerTransport } from '@modelcontextprotocol/sdk/server/streamableHttp.js';

import { ALL_TOOLS } from './tools/index.js';
import { jsonSchemaToZodShape } from './tools/types.js';
import { createClient } from './client.js';

export function createRtylrMcpServer(): McpServer {
  const server = new McpServer(
    { name: 'rtylr-mcp', version: '0.1.0' }
  );

  const api = createClient();

  for (const tool of ALL_TOOLS) {
    server.registerTool(
      tool.name,
      {
        description: tool.description,
        inputSchema: jsonSchemaToZodShape(tool.inputSchema) as any,
      },
      (async (args: Record<string, unknown>) => {
        return tool.handler((args ?? {}) as Record<string, unknown>, api);
      }) as any
    );
  }

  return server;
}

async function main(): Promise<void> {
  const args = process.argv.slice(2);
  const portArg = args.find((a) => a.startsWith('--port='));
  const isPortFlag = args.includes('--port');
  const isHttp = args.includes('--http') || !!portArg || isPortFlag;

  const server = createRtylrMcpServer();

  if (isHttp) {
    let port = 3000;
    if (portArg) {
      port = parseInt(portArg.split('=')[1], 10) || 3000;
    } else if (isPortFlag) {
      const idx = args.indexOf('--port');
      if (idx !== -1 && args[idx + 1]) {
        port = parseInt(args[idx + 1], 10) || 3000;
      }
    }

    const transport = new StreamableHTTPServerTransport({
      sessionIdGenerator: undefined, // Stateless mode per MCP 2.0 (2026-07-28 spec)
    });

    await server.connect(transport);

    const httpServer = createHttpServer((req, res) => {
      transport.handleRequest(req, res);
    });

    httpServer.listen(port, () => {
      process.stderr.write(`rtylr-mcp HTTP server listening statelessly on http://localhost:${port}/mcp\n`);
    });
  } else {
    const transport = new StdioServerTransport();
    await server.connect(transport);
  }
}

main().catch((err) => {
  process.stderr.write(`Fatal: ${err instanceof Error ? err.message : String(err)}\n`);
  process.exit(1);
});
