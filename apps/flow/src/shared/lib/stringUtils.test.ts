import { describe, it, expect } from 'vitest';

import {
  capitalize,
  snakeToTitle,
  camelToSnake,
  snakeToCamel,
  truncate,
  isBlank,
  toSlug,
  initials,
  pluralize,
  countLabel,
  displayDescription,
} from './stringUtils';

describe('capitalize', () => {
  it('capitalizes first character', () => {
    expect(capitalize('hello')).toBe('Hello');
  });

  it('handles empty string', () => {
    expect(capitalize('')).toBe('');
  });
});

describe('snakeToTitle', () => {
  it('converts snake_case to Title Case', () => {
    expect(snakeToTitle('in_progress')).toBe('In Progress');
    expect(snakeToTitle('no_priority')).toBe('No Priority');
  });
});

describe('camelToSnake', () => {
  it('converts camelCase to snake_case', () => {
    expect(camelToSnake('myVariableName')).toBe('my_variable_name');
  });
});

describe('snakeToCamel', () => {
  it('converts snake_case to camelCase', () => {
    expect(snakeToCamel('my_variable_name')).toBe('myVariableName');
  });
});

describe('truncate', () => {
  it('returns string unchanged when under limit', () => {
    expect(truncate('hello', 10)).toBe('hello');
  });

  it('truncates long strings', () => {
    expect(truncate('hello world', 8)).toBe('hello...');
  });

  it('uses custom ellipsis', () => {
    expect(truncate('hello world', 7, '…')).toBe('hello …');
  });
});

describe('isBlank', () => {
  it('returns true for empty string', () => {
    expect(isBlank('')).toBe(true);
  });

  it('returns true for whitespace', () => {
    expect(isBlank('   ')).toBe(true);
  });

  it('returns true for null', () => {
    expect(isBlank(null)).toBe(true);
  });

  it('returns false for non-empty string', () => {
    expect(isBlank('hello')).toBe(false);
  });
});

describe('toSlug', () => {
  it('converts to lowercase slug', () => {
    expect(toSlug('Hello World')).toBe('hello-world');
  });

  it('removes special characters', () => {
    expect(toSlug('My Project! 2026')).toBe('my-project-2026');
  });

  it('trims leading and trailing dashes', () => {
    expect(toSlug('  test  ')).toBe('test');
  });
});

describe('initials', () => {
  it('extracts initials from a name', () => {
    expect(initials('Alice Smith')).toBe('AS');
  });

  it('handles single name', () => {
    expect(initials('Alice')).toBe('A');
  });

  it('caps at two characters', () => {
    expect(initials('Alice Bob Charlie')).toBe('AB');
  });
});

describe('pluralize', () => {
  it('returns singular for count 1', () => {
    expect(pluralize('task', 1)).toBe('task');
  });

  it('returns plural for count 0 and 2+', () => {
    expect(pluralize('task', 0)).toBe('tasks');
    expect(pluralize('task', 5)).toBe('tasks');
  });

  it('uses custom plural form', () => {
    expect(pluralize('person', 2, 'people')).toBe('people');
  });
});

describe('countLabel', () => {
  it('returns singular label for 1', () => {
    expect(countLabel(1, 'task')).toBe('1 task');
  });

  it('returns plural label for other counts', () => {
    expect(countLabel(3, 'task')).toBe('3 tasks');
  });
});

describe('displayDescription', () => {
  it('strips the write-time "-" sentinel to an empty string', () => {
    expect(displayDescription('-')).toBe('');
  });

  it('passes a real description through unchanged', () => {
    expect(displayDescription('Migrate the billing service')).toBe('Migrate the billing service');
  });

  it('treats undefined and null as empty', () => {
    expect(displayDescription(undefined)).toBe('');
    expect(displayDescription(null)).toBe('');
  });

  it('treats an already-empty string as empty', () => {
    expect(displayDescription('')).toBe('');
  });
});
