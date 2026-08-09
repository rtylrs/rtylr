import type { CustomFieldDefinition, Task } from '../stores/flowStore';

/** Find a custom field definition by its ID. */
export function findFieldById(
  fields: CustomFieldDefinition[],
  id: string
): CustomFieldDefinition | undefined {
  return fields.find((f) => f.id === id);
}

/** Returns fields for a specific project, sorted by order. */
export function fieldsForProject(
  fields: CustomFieldDefinition[],
  projectId: string
): CustomFieldDefinition[] {
  return [...fields].filter((f) => f.projectId === projectId).sort((a, b) => a.order - b.order);
}

/** Returns the value of a custom field for a task. */
export function getFieldValue(task: Task, fieldId: string): string | number | undefined {
  return task.customFieldValues?.[fieldId];
}

/** Returns the string representation of a custom field value. */
export function formatFieldValue(
  field: CustomFieldDefinition,
  value: string | number | undefined
): string {
  if (value === undefined || value === null) return '';
  if (field.type === 'number') return String(value);
  if (field.type === 'dropdown') {
    return field.options?.includes(String(value)) ? String(value) : '';
  }
  return String(value);
}

/** Returns true when a dropdown value is valid for the field. */
export function isValidDropdownValue(field: CustomFieldDefinition, value: string): boolean {
  if (field.type !== 'dropdown') return false;
  return !!(field.options && field.options.includes(value));
}

/** Returns the next order number for a new field in a project. */
export function nextFieldOrder(fields: CustomFieldDefinition[], projectId: string): number {
  const projectFields = fieldsForProject(fields, projectId);
  if (projectFields.length === 0) return 0;
  return Math.max(...projectFields.map((f) => f.order)) + 1;
}

/** Returns true if a field name already exists in the project (case-insensitive). */
export function fieldNameExists(
  fields: CustomFieldDefinition[],
  projectId: string,
  name: string
): boolean {
  const lower = name.trim().toLowerCase();
  return fieldsForProject(fields, projectId).some((f) => f.name.toLowerCase() === lower);
}
