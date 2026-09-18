import { describe, it, expect } from 'vitest';
import {
  GLOBAL_FAQS,
  POS_FAQS,
  FLOW_FAQS,
  HR_FAQS,
  CRM_FAQS,
  FINANCE_FAQS,
  ERP_FAQS,
} from './faqs';

describe('GLOBAL_FAQS', () => {
  it('is a non-empty array', () => {
    expect(Array.isArray(GLOBAL_FAQS)).toBe(true);
    expect(GLOBAL_FAQS.length).toBeGreaterThan(0);
  });

  it('every entry has a question string', () => {
    for (const faq of GLOBAL_FAQS) {
      expect(typeof faq.question).toBe('string');
      expect(faq.question.length).toBeGreaterThan(0);
    }
  });

  it('every entry has a non-empty answer string', () => {
    for (const faq of GLOBAL_FAQS) {
      expect(typeof faq.answer).toBe('string');
      expect(faq.answer.length).toBeGreaterThan(0);
    }
  });

  it('contains a question about what rtylr is', () => {
    const found = GLOBAL_FAQS.some((f) => f.question.toLowerCase().includes('what is rtylr'));
    expect(found).toBe(true);
  });

  it('contains a question about pricing', () => {
    const found = GLOBAL_FAQS.some((f) => f.question.toLowerCase().includes('cost'));
    expect(found).toBe(true);
  });
});

describe('POS_FAQS', () => {
  it('is a non-empty array', () => {
    expect(POS_FAQS.length).toBeGreaterThan(0);
  });

  it('contains an offline question', () => {
    const found = POS_FAQS.some((f) => f.question.toLowerCase().includes('offline'));
    expect(found).toBe(true);
  });
});

describe('FLOW_FAQS', () => {
  it('is a non-empty array', () => {
    expect(FLOW_FAQS.length).toBeGreaterThan(0);
  });

  it('contains a question about what Flow is', () => {
    const found = FLOW_FAQS.some((f) => f.question.toLowerCase().includes('flow'));
    expect(found).toBe(true);
  });
});

describe('HR_FAQS', () => {
  it('is a non-empty array', () => {
    expect(HR_FAQS.length).toBeGreaterThan(0);
  });

  it('every entry has both question and answer', () => {
    for (const faq of HR_FAQS) {
      expect(faq.question).toBeTruthy();
      expect(faq.answer).toBeTruthy();
    }
  });
});

describe('CRM_FAQS', () => {
  it('is a non-empty array', () => {
    expect(CRM_FAQS.length).toBeGreaterThan(0);
  });
});

describe('FINANCE_FAQS', () => {
  it('is a non-empty array', () => {
    expect(FINANCE_FAQS.length).toBeGreaterThan(0);
  });

  it('contains a currency question', () => {
    const found = FINANCE_FAQS.some(
      (f) =>
        f.question.toLowerCase().includes('currency') || f.answer.toLowerCase().includes('currency')
    );
    expect(found).toBe(true);
  });
});

describe('ERP_FAQS', () => {
  it('is a non-empty array', () => {
    expect(ERP_FAQS.length).toBeGreaterThan(0);
  });

  it('contains an inventory-related question', () => {
    const found = ERP_FAQS.some(
      (f) =>
        f.question.toLowerCase().includes('inventory') ||
        f.question.toLowerCase().includes('warehouse')
    );
    expect(found).toBe(true);
  });
});

describe('FAQ structure', () => {
  const allFaqSets = [GLOBAL_FAQS, POS_FAQS, FLOW_FAQS, HR_FAQS, CRM_FAQS, FINANCE_FAQS, ERP_FAQS];

  it('no FAQ set is empty', () => {
    for (const set of allFaqSets) {
      expect(set.length).toBeGreaterThan(0);
    }
  });

  it('all questions end with a question mark', () => {
    for (const set of allFaqSets) {
      for (const faq of set) {
        expect(faq.question.trim().endsWith('?')).toBe(true);
      }
    }
  });
});
