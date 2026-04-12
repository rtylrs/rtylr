/**
 * Flow local mode: store-backed reads/writes without the API.
 * By default, production API routes are used.
 * Set VITE_FLOW_LOCAL=true to use local mock data instead.
 */
export function isFlowLocal(): boolean {
  return import.meta.env.VITE_FLOW_LOCAL === 'true';
}
