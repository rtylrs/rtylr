import { describe, it, expect } from 'vitest';

import {
  TASK_STATUSES,
  TASK_PRIORITIES,
  OPEN_STATUSES,
  CLOSED_STATUSES,
  ACTIVE_STATUSES,
  STATUS_LABEL_KEYS,
  DEFAULT_PAGE_SIZE,
  MAX_TASK_TITLE_LENGTH,
  MAX_PROJECT_NAME_LENGTH,
  EFFORT_BUCKETS_MINUTES,
  DEFAULT_CYCLE_DAYS,
  PROJECT_HEALTH_OPTIONS,
  MILESTONE_STATUSES,
} from './constantsFlow';

describe('TASK_STATUSES', () => {
  it('contains all expected statuses', () => {
    expect(TASK_STATUSES).toContain('todo');
    expect(TASK_STATUSES).toContain('in_progress');
    expect(TASK_STATUSES).toContain('done');
    expect(TASK_STATUSES).toContain('cancelled');
    expect(TASK_STATUSES).toContain('backlog');
    expect(TASK_STATUSES).toContain('in_review');
  });

  it('has 6 statuses', () => {
    expect(TASK_STATUSES.length).toBe(6);
  });
});

describe('TASK_PRIORITIES', () => {
  it('contains all priority levels', () => {
    expect(TASK_PRIORITIES).toContain('urgent');
    expect(TASK_PRIORITIES).toContain('no_priority');
  });

  it('has 5 priorities', () => {
    expect(TASK_PRIORITIES.length).toBe(5);
  });
});

describe('OPEN_STATUSES', () => {
  it('excludes done and cancelled', () => {
    expect(OPEN_STATUSES).not.toContain('done');
    expect(OPEN_STATUSES).not.toContain('cancelled');
  });
});

describe('CLOSED_STATUSES', () => {
  it('includes done and cancelled', () => {
    expect(CLOSED_STATUSES).toContain('done');
    expect(CLOSED_STATUSES).toContain('cancelled');
  });

  it('has exactly 2 statuses', () => {
    expect(CLOSED_STATUSES.length).toBe(2);
  });
});

describe('ACTIVE_STATUSES', () => {
  it('includes in_progress and in_review', () => {
    expect(ACTIVE_STATUSES).toContain('in_progress');
    expect(ACTIVE_STATUSES).toContain('in_review');
  });
});

describe('STATUS_LABEL_KEYS', () => {
  it('has an i18n key for every task status', () => {
    for (const status of TASK_STATUSES) {
      expect(STATUS_LABEL_KEYS[status]).toBeTruthy();
    }
  });

  it('resolves the "todo" status to the same i18n key used by the task detail status picker', () => {
    // Regression guard for issue #756: the project List view and the task
    // detail status picker each used to hardcode their own "todo" label
    // ("Todo" vs "To Do") instead of sharing this map. Anyone reading a
    // status label must go through STATUS_LABEL_KEYS so the two can't
    // drift apart again.
    expect(STATUS_LABEL_KEYS.todo).toBe('flow.taskDetail.todo');
  });

  it('points every status at the flow.taskDetail namespace', () => {
    for (const status of TASK_STATUSES) {
      expect(STATUS_LABEL_KEYS[status]).toMatch(/^flow\.taskDetail\./);
    }
  });
});

describe('DEFAULT_PAGE_SIZE', () => {
  it('is a positive number', () => {
    expect(DEFAULT_PAGE_SIZE).toBeGreaterThan(0);
  });
});

describe('MAX_TASK_TITLE_LENGTH', () => {
  it('is 255', () => {
    expect(MAX_TASK_TITLE_LENGTH).toBe(255);
  });
});

describe('MAX_PROJECT_NAME_LENGTH', () => {
  it('is 100', () => {
    expect(MAX_PROJECT_NAME_LENGTH).toBe(100);
  });
});

describe('EFFORT_BUCKETS_MINUTES', () => {
  it('contains quick effort sizes', () => {
    expect(EFFORT_BUCKETS_MINUTES).toContain(15);
    expect(EFFORT_BUCKETS_MINUTES).toContain(60);
  });
});

describe('DEFAULT_CYCLE_DAYS', () => {
  it('is 14 days', () => {
    expect(DEFAULT_CYCLE_DAYS).toBe(14);
  });
});

describe('PROJECT_HEALTH_OPTIONS', () => {
  it('contains all health statuses', () => {
    expect(PROJECT_HEALTH_OPTIONS).toContain('on_track');
    expect(PROJECT_HEALTH_OPTIONS).toContain('at_risk');
    expect(PROJECT_HEALTH_OPTIONS).toContain('off_track');
  });
});

describe('MILESTONE_STATUSES', () => {
  it('contains all milestone statuses', () => {
    expect(MILESTONE_STATUSES).toContain('planning');
    expect(MILESTONE_STATUSES).toContain('completed');
  });
});
