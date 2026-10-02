import type { AxiosInstance } from 'axios';
import { z } from 'zod';

export interface ToolResult {
  content: Array<{ type: 'text'; text: string }>;
  isError?: boolean;
  [key: string]: unknown;
}

export interface Tool {
  name: string;
  description: string;
  inputSchema: {
    type: 'object';
    properties: Record<string, { type: string; description?: string; enum?: string[] }>;
    required?: string[];
  };
  handler: (input: Record<string, unknown>, api: AxiosInstance) => Promise<ToolResult>;
}

export function ok(data: unknown): ToolResult {
  return { content: [{ type: 'text', text: JSON.stringify(data, null, 2) }] };
}

export function err(message: string): ToolResult {
  return { content: [{ type: 'text', text: `Error: ${message}` }], isError: true };
}

export function apiErr(e: unknown): ToolResult {
  if (e && typeof e === 'object' && 'response' in e) {
    const axiosErr = e as { response?: { data?: { message?: string } }; message: string };
    return err(axiosErr.response?.data?.message ?? axiosErr.message);
  }
  return err(String(e));
}

export function qs(input: Record<string, unknown>, keys: string[]): string {
  const p = new URLSearchParams();
  for (const k of keys) {
    if (input[k] != null) p.set(k, String(input[k]));
  }
  const s = p.toString();
  return s ? `?${s}` : '';
}

export function jsonSchemaToZodShape(inputSchema: {
  type?: string;
  properties?: Record<string, { type?: string; description?: string; enum?: string[] }>;
  required?: string[];
}): Record<string, z.ZodTypeAny> {
  const shape: Record<string, z.ZodTypeAny> = {};
  const props = inputSchema?.properties ?? {};
  const required = new Set(inputSchema?.required ?? []);

  for (const [key, prop] of Object.entries(props)) {
    let schema: z.ZodTypeAny;

    if (prop.enum && prop.enum.length > 0) {
      schema = z.enum(prop.enum as [string, ...string[]]);
    } else if (prop.type === 'integer') {
      schema = z.number().int();
    } else if (prop.type === 'number') {
      schema = z.number();
    } else if (prop.type === 'boolean') {
      schema = z.boolean();
    } else if (prop.type === 'array') {
      schema = z.array(z.any());
    } else if (prop.type === 'object') {
      schema = z.record(z.any());
    } else {
      schema = z.string();
    }

    if (prop.description) {
      schema = schema.describe(prop.description);
    }

    if (!required.has(key)) {
      schema = schema.optional();
    }

    shape[key] = schema;
  }

  return shape;
}
