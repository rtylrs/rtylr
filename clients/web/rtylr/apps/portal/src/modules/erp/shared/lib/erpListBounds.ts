export const ERP_LIST_LIMIT = 100;
export const ERP_LIST_OFFSET = 0;

export function withErpListBounds<T extends Record<string, unknown>>(
  request: T
): T & { limit: number; offset: number } {
  return {
    ...request,
    limit: ERP_LIST_LIMIT,
    offset: ERP_LIST_OFFSET,
  };
}
