import { describe, it, expect } from 'vitest';
import {
  SUBMISSION_STATUSES,
  SUBMISSION_STATUS_LABELS,
  SUBMISSION_STATUS_COLORS,
  SENIORITY_LEVELS,
  SENIORITY_LABELS,
  WORK_MODES,
  WORK_MODE_LABELS,
  EMPLOYMENT_TYPES,
  TEMPLATE_IDS,
  MAX_RESUME_SIZE_BYTES,
  ACCEPTED_RESUME_TYPES,
} from './constants';

describe('SUBMISSION_STATUSES', () => {
  it('includes all 5 statuses', () => {
    expect(SUBMISSION_STATUSES).toHaveLength(5);
    expect(SUBMISSION_STATUSES).toContain('pending');
    expect(SUBMISSION_STATUSES).toContain('hired');
  });
});

describe('SUBMISSION_STATUS_LABELS', () => {
  it('has a label for every status', () => {
    for (const status of SUBMISSION_STATUSES) {
      expect(SUBMISSION_STATUS_LABELS[status]).toBeTruthy();
    }
  });

  it('Pending label is capitalized', () => {
    expect(SUBMISSION_STATUS_LABELS['pending']).toBe('Pending');
  });
});

describe('SUBMISSION_STATUS_COLORS', () => {
  it('has a color for every status', () => {
    for (const status of SUBMISSION_STATUSES) {
      expect(SUBMISSION_STATUS_COLORS[status]).toMatch(/^#[0-9a-f]{6}$/i);
    }
  });

  it('hired uses green', () => {
    expect(SUBMISSION_STATUS_COLORS['hired']).toBe('#10b981');
  });

  it('rejected uses red', () => {
    expect(SUBMISSION_STATUS_COLORS['rejected']).toBe('#ef4444');
  });
});

describe('SENIORITY_LEVELS', () => {
  it('starts with intern and ends with executive', () => {
    expect(SENIORITY_LEVELS[0]).toBe('intern');
    expect(SENIORITY_LEVELS[SENIORITY_LEVELS.length - 1]).toBe('executive');
  });

  it('contains senior', () => {
    expect(SENIORITY_LEVELS).toContain('senior');
  });
});

describe('SENIORITY_LABELS', () => {
  it('has a label for every seniority level', () => {
    for (const level of SENIORITY_LEVELS) {
      expect(SENIORITY_LABELS[level]).toBeTruthy();
    }
  });

  it('mid-level is labeled Mid-level', () => {
    expect(SENIORITY_LABELS['mid']).toBe('Mid-level');
  });
});

describe('WORK_MODES', () => {
  it('contains remote, hybrid, and onsite', () => {
    expect(WORK_MODES).toContain('remote');
    expect(WORK_MODES).toContain('hybrid');
    expect(WORK_MODES).toContain('onsite');
  });
});

describe('WORK_MODE_LABELS', () => {
  it('onsite is labeled On-site', () => {
    expect(WORK_MODE_LABELS['onsite']).toBe('On-site');
  });
});

describe('EMPLOYMENT_TYPES', () => {
  it('contains at least full-time and contract', () => {
    expect(EMPLOYMENT_TYPES).toContain('full-time');
    expect(EMPLOYMENT_TYPES).toContain('contract');
  });
});

describe('TEMPLATE_IDS', () => {
  it('contains minimal, editorial, and bold', () => {
    expect(TEMPLATE_IDS).toContain('minimal');
    expect(TEMPLATE_IDS).toContain('editorial');
    expect(TEMPLATE_IDS).toContain('bold');
  });
});

describe('MAX_RESUME_SIZE_BYTES', () => {
  it('is 5 MB', () => {
    expect(MAX_RESUME_SIZE_BYTES).toBe(5 * 1024 * 1024);
  });
});

describe('ACCEPTED_RESUME_TYPES', () => {
  it('includes PDF', () => {
    expect(ACCEPTED_RESUME_TYPES).toContain('application/pdf');
  });

  it('includes at least 2 types', () => {
    expect(ACCEPTED_RESUME_TYPES.length).toBeGreaterThanOrEqual(2);
  });
});
