import { describe, it, expect } from 'vitest';
import {
  fadeInUp,
  fadeInDown,
  fadeIn,
  scaleIn,
  slideInLeft,
  slideInRight,
  staggerContainer,
  staggerContainerSlow,
  blurFadeIn,
  blurFadeInSlow,
  staggerBlur,
  viewportOnce,
} from './animations';

describe('fadeInUp', () => {
  it('has a hidden state with zero opacity and positive y offset', () => {
    expect(fadeInUp.hidden).toMatchObject({ opacity: 0, y: 30 });
  });

  it('has a visible state with full opacity and zero y', () => {
    const visible = fadeInUp.visible as Record<string, unknown>;
    expect(visible.opacity).toBe(1);
    expect(visible.y).toBe(0);
  });
});

describe('fadeInDown', () => {
  it('has a hidden state with negative y offset', () => {
    expect(fadeInDown.hidden).toMatchObject({ opacity: 0, y: -30 });
  });

  it('becomes fully visible at y 0', () => {
    const visible = fadeInDown.visible as Record<string, unknown>;
    expect(visible.opacity).toBe(1);
    expect(visible.y).toBe(0);
  });
});

describe('fadeIn', () => {
  it('starts with zero opacity', () => {
    expect(fadeIn.hidden).toMatchObject({ opacity: 0 });
  });

  it('ends with full opacity', () => {
    const visible = fadeIn.visible as Record<string, unknown>;
    expect(visible.opacity).toBe(1);
  });

  it('does not animate position', () => {
    expect((fadeIn.hidden as Record<string, unknown>).y).toBeUndefined();
  });
});

describe('scaleIn', () => {
  it('starts scaled down', () => {
    expect(scaleIn.hidden).toMatchObject({ opacity: 0, scale: 0.9 });
  });

  it('ends at scale 1', () => {
    const visible = scaleIn.visible as Record<string, unknown>;
    expect(visible.scale).toBe(1);
    expect(visible.opacity).toBe(1);
  });
});

describe('slideInLeft', () => {
  it('starts with a negative x offset', () => {
    expect((slideInLeft.hidden as Record<string, unknown>).x).toBeLessThan(0);
  });

  it('ends at x 0', () => {
    const visible = slideInLeft.visible as Record<string, unknown>;
    expect(visible.x).toBe(0);
  });
});

describe('slideInRight', () => {
  it('starts with a positive x offset', () => {
    expect((slideInRight.hidden as Record<string, unknown>).x).toBeGreaterThan(0);
  });

  it('ends at x 0', () => {
    const visible = slideInRight.visible as Record<string, unknown>;
    expect(visible.x).toBe(0);
  });
});

describe('staggerContainer', () => {
  it('has a staggerChildren transition', () => {
    const visible = staggerContainer.visible as { transition: Record<string, unknown> };
    expect(visible.transition.staggerChildren).toBeDefined();
  });
});

describe('staggerContainerSlow', () => {
  it('has a longer stagger delay than staggerContainer', () => {
    const fast = (staggerContainer.visible as { transition: Record<string, unknown> }).transition;
    const slow = (staggerContainerSlow.visible as { transition: Record<string, unknown> })
      .transition;
    expect(Number(slow.staggerChildren)).toBeGreaterThan(Number(fast.staggerChildren));
  });
});

describe('blurFadeIn', () => {
  it('starts blurred', () => {
    const hidden = blurFadeIn.hidden as Record<string, unknown>;
    expect(typeof hidden.filter).toBe('string');
    expect(String(hidden.filter)).toContain('blur');
  });

  it('ends with no blur', () => {
    const visible = blurFadeIn.visible as Record<string, unknown>;
    expect(visible.filter).toBe('blur(0px)');
  });
});

describe('blurFadeInSlow', () => {
  it('starts with more blur than blurFadeIn', () => {
    const slowHidden = blurFadeInSlow.hidden as Record<string, unknown>;
    const fastHidden = blurFadeIn.hidden as Record<string, unknown>;
    const slowBlur = parseFloat(String(slowHidden.filter).replace(/[^0-9.]/g, ''));
    const fastBlur = parseFloat(String(fastHidden.filter).replace(/[^0-9.]/g, ''));
    expect(slowBlur).toBeGreaterThan(fastBlur);
  });
});

describe('staggerBlur', () => {
  it('has staggerChildren set', () => {
    const visible = staggerBlur.visible as { transition: Record<string, unknown> };
    expect(visible.transition.staggerChildren).toBeGreaterThan(0);
  });
});

describe('viewportOnce', () => {
  it('fires only once', () => {
    expect(viewportOnce.once).toBe(true);
  });

  it('has a negative margin to trigger early', () => {
    expect(viewportOnce.margin).toMatch(/^-\d/);
  });
});
