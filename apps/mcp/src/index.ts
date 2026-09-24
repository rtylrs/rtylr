#!/usr/bin/env node
import {
  createServer as createHttpServer,
  type IncomingMessage,
  type ServerResponse,
} from "node:http";
import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";

import { ALL_TOOLS } from "./tools/index.js";
import { jsonSchemaToZodShape } from "./tools/types.js";
import { createClient } from "./client.js";

const MAX_HTTP_BODY_BYTES = 1024 * 1024;

function resolveVersion(): string {
  try {
    const pkgPath = join(
      dirname(fileURLToPath(import.meta.url)),
      "../package.json",
    );
    const pkg = JSON.parse(readFileSync(pkgPath, "utf8")) as {
      version?: string;
    };
    return pkg.version ?? "0.0.0";
  } catch {
    return "0.0.0";
  }
}

export function createRtylrMcpServer(): McpServer {
  const server = new McpServer({
    name: "rtylr-mcp",
    version: resolveVersion(),
  });

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
      }) as any,
    );
  }

  return server;
}

export function parsePort(value: string | undefined): number | null {
  if (!value || !/^\d+$/.test(value)) return null;
  const port = Number(value);
  if (!Number.isInteger(port) || port < 1 || port > 65535) {
    return null;
  }
  return port;
}

function extractPort(args: string[]): number {
  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg.startsWith("--port=")) {
      const port = parsePort(arg.slice("--port=".length));
      if (port === null) {
        process.stderr.write(`Invalid port: ${arg}\n`);
        process.exit(2);
      }
      return port;
    }
    if (arg === "--port") {
      const port = parsePort(args[i + 1]);
      if (port === null) {
        process.stderr.write(`Invalid port: ${args[i + 1] ?? ""}\n`);
        process.exit(2);
      }
      return port;
    }
  }
  return 3000;
}

type BodyReadResult =
  { ok: true; bytes: Buffer } | { ok: false; tooLarge: true };

function sendJsonError(
  res: ServerResponse,
  status: number,
  error: string,
): void {
  if (res.headersSent) {
    res.end();
    return;
  }
  res.writeHead(status, { "Content-Type": "application/json" });
  res.end(JSON.stringify({ error }));
}

function sendParseError(res: ServerResponse): void {
  if (res.headersSent) {
    res.end();
    return;
  }
  res.writeHead(400, { "Content-Type": "application/json" });
  res.end(
    JSON.stringify({
      jsonrpc: "2.0",
      error: { code: -32700, message: "Parse error: Invalid JSON" },
      id: null,
    }),
  );
}

function sendPayloadTooLarge(res: ServerResponse): void {
  if (res.headersSent) {
    res.end();
    return;
  }
  res.writeHead(413, {
    "Content-Type": "application/json",
    Connection: "close",
  });
  res.end(JSON.stringify({ error: "payload_too_large" }));
}

function readRequestBody(
  req: IncomingMessage,
  res: ServerResponse,
): Promise<BodyReadResult> {
  const declaredLength = req.headers["content-length"];
  const rawLength = Array.isArray(declaredLength)
    ? declaredLength[0]
    : declaredLength;
  if (rawLength !== undefined) {
    const length = Number(rawLength);
    if (Number.isFinite(length) && length > MAX_HTTP_BODY_BYTES) {
      req.resume();
      sendPayloadTooLarge(res);
      return Promise.resolve({ ok: false, tooLarge: true });
    }
  }

  return new Promise<BodyReadResult>((resolve, reject) => {
    const chunks: Buffer[] = [];
    let total = 0;
    let settled = false;

    const rejectTooLarge = (): void => {
      if (settled) return;
      settled = true;
      chunks.length = 0;
      req.resume();
      sendPayloadTooLarge(res);
      resolve({ ok: false, tooLarge: true });
    };

    req.on("data", (chunk: Buffer | string) => {
      if (settled) return;
      const bytes = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk);
      total += bytes.length;
      if (total > MAX_HTTP_BODY_BYTES) {
        rejectTooLarge();
        return;
      }
      chunks.push(bytes);
    });

    req.once("end", () => {
      if (settled) return;
      settled = true;
      resolve({ ok: true, bytes: Buffer.concat(chunks) });
    });

    req.once("error", (error) => {
      if (settled) return;
      settled = true;
      reject(error);
    });
  });
}

async function handleHttpRequest(
  req: IncomingMessage,
  res: ServerResponse,
  port: number,
): Promise<void> {
  const pathname = req.url?.split("?", 1)[0] ?? "";
  if (pathname !== "/mcp") {
    sendJsonError(res, 404, "not_found");
    return;
  }

  let parsedBody: unknown;
  if (req.method === "POST") {
    let body: BodyReadResult;
    try {
      body = await readRequestBody(req, res);
    } catch (error) {
      process.stderr.write(
        `Error reading request: ${error instanceof Error ? error.message : String(error)}\n`,
      );
      sendParseError(res);
      return;
    }
    if (!body.ok) return;

    try {
      parsedBody = JSON.parse(body.bytes.toString("utf8"));
    } catch {
      sendParseError(res);
      return;
    }
  }

  const requestServer = createRtylrMcpServer();
  const transport = new StreamableHTTPServerTransport({
    sessionIdGenerator: undefined,
    enableJsonResponse: true,
    enableDnsRebindingProtection: true,
    allowedHosts: [`127.0.0.1:${port}`, `localhost:${port}`],
    allowedOrigins: [`http://127.0.0.1:${port}`, `http://localhost:${port}`],
  });
  let cleanedUp = false;
  const cleanup = async (): Promise<void> => {
    if (cleanedUp) return;
    cleanedUp = true;
    try {
      await requestServer.close();
    } catch {
      // The request may have failed before the server finished connecting.
    }
    try {
      await transport.close();
    } catch {
      // The transport may already have closed itself.
    }
  };
  const cleanupOnResponse = (): void => {
    void cleanup();
  };
  res.once("finish", cleanupOnResponse);
  res.once("close", cleanupOnResponse);

  try {
    await requestServer.connect(transport);
    if (req.method === "POST") {
      await transport.handleRequest(req, res, parsedBody);
    } else {
      await transport.handleRequest(req, res);
    }
    if (req.method !== "GET") await cleanup();
  } catch (err) {
    process.stderr.write(
      `Error handling request: ${err instanceof Error ? err.message : String(err)}\n`,
    );
    if (!res.headersSent) {
      res.writeHead(500, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ error: "internal_error" }));
    } else {
      res.end();
    }
    await cleanup();
  }
}

async function main(): Promise<void> {
  const args = process.argv.slice(2);
  const isHttp =
    args.includes("--http") ||
    args.some((a) => a === "--port" || a.startsWith("--port="));

  if (isHttp) {
    const port = extractPort(args);

    const httpServer = createHttpServer((req, res) => {
      void handleHttpRequest(req, res, port).catch((err) => {
        process.stderr.write(
          `Error handling request: ${err instanceof Error ? err.message : String(err)}\n`,
        );
        if (!res.headersSent) {
          res.writeHead(500, { "Content-Type": "application/json" });
          res.end(JSON.stringify({ error: "internal_error" }));
        } else {
          res.end();
        }
      });
    });

    httpServer.on("error", (err) => {
      const msg =
        err instanceof Error && "code" in err && err.code === "EADDRINUSE"
          ? `Port ${port} is already in use`
          : err instanceof Error
            ? err.message
            : String(err);
      process.stderr.write(`Fatal: ${msg}\n`);
      process.exit(1);
    });

    httpServer.listen(port, "127.0.0.1", () => {
      process.stderr.write(
        `rtylr-mcp HTTP server listening statelessly on http://127.0.0.1:${port}/mcp\n`,
      );
    });
  } else {
    const server = createRtylrMcpServer();
    const transport = new StdioServerTransport();
    await server.connect(transport);
  }
}

main().catch((err) => {
  process.stderr.write(
    `Fatal: ${err instanceof Error ? err.message : String(err)}\n`,
  );
  process.exit(1);
});
