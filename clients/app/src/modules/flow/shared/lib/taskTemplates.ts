import type { TFunction } from 'i18next';
import type { ApiTaskTemplate } from '../services/flowApi';

/**
 * The description a new task of `type` starts with: the project's own template
 * for the type, else the built-in bug template for a bug, else nothing.
 */
export function templateFor(
  type: string,
  templates: readonly ApiTaskTemplate[],
  t: TFunction
): string {
  const own = templates.find((template) => template.taskType === type)?.body;
  if (own) return own;
  return type === 'bug' ? t('flow.taskTemplates.defaultBug') : '';
}

const BRANCH_PREFIX: Record<string, string> = {
  bug: 'fix',
  feature: 'feat',
  chore: 'chore',
  spike: 'spike',
  epic: 'epic',
};

/**
 * A git branch name for the task, prefixed by its type: fix/t-42-login-button.
 * Arabic and other non-Latin titles keep only the key, since git refs with
 * them are awkward to type.
 */
export function branchNameFor(task: {
  identifier?: string;
  title: string;
  taskType?: string;
}): string {
  const prefix = BRANCH_PREFIX[task.taskType ?? ''] ?? 'task';
  const key = (task.identifier ?? '').toLowerCase().replace(/[^a-z0-9-]+/g, '');
  const slug = task.title
    .normalize('NFKD')
    .replace(/[̀-ͯ]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .split('-')
    .filter(Boolean)
    .slice(0, 6)
    .join('-');
  return `${prefix}/${[key, slug].filter(Boolean).join('-')}`;
}
