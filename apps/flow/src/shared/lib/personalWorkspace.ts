/** Stable id for the default Personal workspace (always present; cannot be deleted). */
export const PERSONAL_WORKSPACE_ID = 'personal';

/**
 * True when an id is safe to send to the API as `workspace_id`.
 *
 * `PERSONAL_WORKSPACE_ID` is a client-only sentinel so the layout can paint
 * before the workspace fetch resolves - it is not a real record. Real ids
 * arrive from the wire as int64 and reach the store stringified, so anything
 * that is not a positive whole number is unresolved local state.
 *
 * A bare `!!workspaceId` check passes the sentinel through and produces a
 * guaranteed 400: the contracts declare `workspace_id: int64 @required @gt(0)`
 * (see packages/voxie/api/flow/dashboard/v1/models.onk and
 * flow/project/v1/models.onk), so "personal" and "0" are both rejected.
 */
export function isApiWorkspaceId(id: string | undefined): boolean {
  if (!id) return false;
  const numeric = Number(id);
  return Number.isInteger(numeric) && numeric > 0;
}
