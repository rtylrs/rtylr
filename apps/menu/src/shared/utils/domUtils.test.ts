import { describe, it, expect, vi, beforeEach } from 'vitest';
import { scrollToId, dataAttrFromClick, isScrolledToBottom } from './domUtils';

describe('scrollToId', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
  });

  it('calls scrollIntoView on the target element', () => {
    const el = document.createElement('div');
    el.id = 'target';
    const spy = vi.spyOn(el, 'scrollIntoView');
    document.body.appendChild(el);

    scrollToId('target');

    expect(spy).toHaveBeenCalledWith({ behavior: 'smooth', block: 'start' });
  });

  it('does nothing when the element is not found', () => {
    expect(() => scrollToId('nonexistent')).not.toThrow();
  });

  it('passes the behavior argument through', () => {
    const el = document.createElement('div');
    el.id = 'el';
    const spy = vi.spyOn(el, 'scrollIntoView');
    document.body.appendChild(el);

    scrollToId('el', 'instant');

    expect(spy).toHaveBeenCalledWith({ behavior: 'instant', block: 'start' });
  });
});

describe('dataAttrFromClick', () => {
  it('returns the data attribute from the direct target', () => {
    const el = document.createElement('button');
    el.dataset.action = 'add';
    const event = { target: el } as unknown as MouseEvent;

    expect(dataAttrFromClick(event, 'action')).toBe('add');
  });

  it('walks up the DOM tree to find the attribute', () => {
    const parent = document.createElement('div');
    parent.dataset.itemId = '42';
    const child = document.createElement('span');
    parent.appendChild(child);
    const event = { target: child } as unknown as MouseEvent;

    expect(dataAttrFromClick(event, 'itemId')).toBe('42');
  });

  it('returns null when the attribute is not found', () => {
    const el = document.createElement('button');
    const event = { target: el } as unknown as MouseEvent;

    expect(dataAttrFromClick(event, 'missing')).toBeNull();
  });
});

describe('isScrolledToBottom', () => {
  it('returns true when scrolled to the bottom', () => {
    const el = {
      scrollHeight: 500,
      scrollTop: 490,
      clientHeight: 10,
    } as HTMLElement;

    expect(isScrolledToBottom(el)).toBe(true);
  });

  it('returns false when not scrolled to the bottom', () => {
    const el = {
      scrollHeight: 500,
      scrollTop: 100,
      clientHeight: 10,
    } as HTMLElement;

    expect(isScrolledToBottom(el)).toBe(false);
  });

  it('accepts a custom threshold', () => {
    const el = {
      scrollHeight: 500,
      scrollTop: 480,
      clientHeight: 10,
    } as HTMLElement;

    expect(isScrolledToBottom(el, 20)).toBe(true);
    expect(isScrolledToBottom(el, 5)).toBe(false);
  });
});
