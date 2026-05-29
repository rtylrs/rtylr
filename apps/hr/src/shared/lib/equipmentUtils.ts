export type EquipmentStatus = 'Available' | 'Assigned' | 'In Repair' | 'Retired';
export type EquipmentCategory = 'Laptop' | 'Phone' | 'Monitor' | 'Desk' | 'Chair' | 'Other';

export function equipmentStatusLabel(status: EquipmentStatus): string {
  const labels: Record<EquipmentStatus, string> = {
    Available: 'Available',
    Assigned: 'Assigned',
    'In Repair': 'In Repair',
    Retired: 'Retired',
  };
  return labels[status] ?? status;
}

export function equipmentStatusColor(status: EquipmentStatus): string {
  const colors: Record<EquipmentStatus, string> = {
    Available: 'green',
    Assigned: 'blue',
    'In Repair': 'yellow',
    Retired: 'gray',
  };
  return colors[status] ?? 'gray';
}

export function isAssignable(status: EquipmentStatus): boolean {
  return status === 'Available';
}

export function isAssigned(status: EquipmentStatus): boolean {
  return status === 'Assigned';
}

export function isActiveEquipment(status: EquipmentStatus): boolean {
  return status === 'Available' || status === 'Assigned';
}

export function depreciatedValue(
  originalCost: number,
  ageYears: number,
  usefulLifeYears: number
): number {
  if (usefulLifeYears <= 0) return 0;
  const remaining = Math.max(0, usefulLifeYears - ageYears);
  return Math.round(((originalCost * remaining) / usefulLifeYears) * 100) / 100;
}

export function countAssigned(records: Array<{ status: EquipmentStatus }>): number {
  return records.filter((r) => r.status === 'Assigned').length;
}

export function countAvailable(records: Array<{ status: EquipmentStatus }>): number {
  return records.filter((r) => r.status === 'Available').length;
}

export function filterByEquipmentCategory<T extends { category: EquipmentCategory }>(
  records: T[],
  category: EquipmentCategory
): T[] {
  return records.filter((r) => r.category === category);
}

export function utilizationRate(records: Array<{ status: EquipmentStatus }>): number {
  const active = records.filter((r) => isActiveEquipment(r.status)).length;
  if (active === 0) return 0;
  const assigned = records.filter((r) => r.status === 'Assigned').length;
  return Math.round((assigned / active) * 100) / 100;
}
