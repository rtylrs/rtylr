interface CartContext {
  branchId: number;
  terminalId: number;
  employeeId: number;
}

function normalizePositiveId(id: unknown): number | null {
  const value = typeof id === 'string' ? Number(id) : id;
  if (typeof value !== 'number' || !Number.isFinite(value) || value <= 0) return null;
  return Math.floor(value);
}

export function buildCartContext(
  locationId: number | null,
  terminalId: number | null,
  employeeId: unknown
): CartContext | null {
  const branch = normalizePositiveId(locationId);
  const terminal = normalizePositiveId(terminalId);
  const employee = normalizePositiveId(employeeId);

  if (!branch || !terminal || !employee) return null;
  return { branchId: branch, terminalId: terminal, employeeId: employee };
}
