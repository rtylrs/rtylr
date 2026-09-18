#!/usr/bin/env node
import { createServer as createHttpServer } from 'node:http';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { StreamableHTTPServerTransport } from '@modelcontextprotocol/sdk/server/streamableHttp.js';

import { ALL_TOOLS } from './tools/index.js';
import { jsonSchemaToZodShape } from './tools/types.js';
import { createClient } from './client.js';

function resolveVersion(): string {
  try {
    const pkgPath = join(dirname(fileURLToPath(import.meta.url)), '../package.json');
    const pkg = JSON.parse(readFileSync(pkgPath, 'utf8')) as { version?: string };
    return pkg.version ?? '0.0.0';
  } catch {
    return '0.0.0';
  }
}

export function createRtylrMcpServer(): McpServer {
  const server = new McpServer(
    { name: 'rtylr-mcp', version: resolveVersion() }
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

function parsePort(value: string | undefined): number | null {
  if (!value) return null;
  const port = Number.parseInt(value, 10);
  if (!Number.isInteger(port) || port < 1 || port > 65535) {
    return null;
  }
  return port;
}

function extractPort(args: string[]): number {
  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg.startsWith('--port=')) {
      const port = parsePort(arg.slice('--port='.length));
      if (port === null) {
        process.stderr.write(`Invalid port: ${arg}\n`);
        process.exit(2);
      }
      return port;
    }
    if (arg === '--port') {
      const port = parsePort(args[i + 1]);
      if (port === null) {
        process.stderr.write(`Invalid port: ${args[i + 1] ?? ''}\n`);
        process.exit(2);
      }
      return port;
    }
  }
  return 3000;
}

async function main(): Promise<void> {
  const args = process.argv.slice(2);
  const isHttp = args.includes('--http') || args.some((a) => a === '--port' || a.startsWith('--port='));

  const server = createRtylrMcpServer();

  if (isHttp) {
    const port = extractPort(args);

    const transport = new StreamableHTTPServerTransport({
      sessionIdGenerator: undefined, // Stateless mode per MCP 2.0 (2026-07-28 spec)
    });

    await server.connect(transport);

    const httpServer = createHttpServer((req, res) => {
      transport.handleRequest(req, res).catch((err) => {
        process.stderr.write(`Error handling request: ${err instanceof Error ? err.message : String(err)}\n`);
        if (!res.headersSent) {
          res.writeHead(500, { 'Content-Type': 'application/json' });
          res.end(JSON.stringify({ error: 'internal_error' }));
        } else {
          res.end();
        }
      });
    });

    httpServer.on('error', (err) => {
      const msg = err instanceof Error && 'code' in err && err.code === 'EADDRINUSE'
        ? `Port ${port} is already in use`
        : err instanceof Error ? err.message : String(err);
      process.stderr.write(`Fatal: ${msg}\n`);
      process.exit(1);
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
