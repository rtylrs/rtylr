import type { AxiosInstance } from 'axios';

export interface ToolResult {
  content: Array<{ type: 'text'; text: string }>;
  isError?: boolean;
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
