/**
 * Local-first Flow: rich demo data and store-backed reads/writes without the API.
 * Set VITE_FLOW_LOCAL=false to use production API routes instead.
 */
export function isFlowLocal(): boolean {
  return import.meta.env.VITE_FLOW_LOCAL !== 'false';
}
