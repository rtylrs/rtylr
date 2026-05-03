import { homedir } from 'os';
import { join } from 'path';
import { readFileSync, writeFileSync, mkdirSync, existsSync } from 'fs';

export const API_BASE = process.env.RTYLR_API_URL ?? 'https://api.voxire.dev';
export const OAUTH_CLIENT_ID = 'rtylr-mcp';
export const OAUTH_SCOPES = 'flow:read flow:write crm:read crm:write finance:read finance:write erp:read erp:write';

const CONFIG_DIR = join(homedir(), '.rtylr-mcp');
const CONFIG_FILE = join(CONFIG_DIR, 'config.json');

export interface StoredConfig {
  access_token: string;
  refresh_token: string;
  expires_at: number;
}

export function loadConfig(): StoredConfig | null {
  try {
    if (!existsSync(CONFIG_FILE)) return null;
    const raw = readFileSync(CONFIG_FILE, 'utf8');
    return JSON.parse(raw) as StoredConfig;
  } catch {
    return null;
  }
}

export function saveConfig(config: StoredConfig): void {
  if (!existsSync(CONFIG_DIR)) {
    mkdirSync(CONFIG_DIR, { recursive: true, mode: 0o700 });
  }
  writeFileSync(CONFIG_FILE, JSON.stringify(config, null, 2), { mode: 0o600 });
}

export function clearConfig(): void {
  try {
    if (existsSync(CONFIG_FILE)) {
      writeFileSync(CONFIG_FILE, '{}');
    }
  } catch {
    // ignore
  }
}
