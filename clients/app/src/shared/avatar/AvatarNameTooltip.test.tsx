import React from 'react';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { Avatar } from '@voxire/ui';
import { AvatarNameTooltip } from './AvatarNameTooltip';
import { AssigneeAvatar } from '../../modules/flow/shared/components/AssigneeAvatar';

function hover(el: Element, pointerType = 'mouse'): void {
  fireEvent.pointerOver(el, { pointerType });
}

describe('AvatarNameTooltip', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows the name after a short hover and hides it on leave', () => {
    render(
      <>
        <Avatar initials="SH" seed="7" name="Sam Hill" />
        <AvatarNameTooltip />
      </>
    );
    const avatar = screen.getByRole('img', { name: 'Sam Hill' });
    hover(avatar);
    expect(screen.queryByRole('tooltip')).toBeNull();
    act(() => {
      vi.advanceTimersByTime(200);
    });
    expect(screen.getByRole('tooltip').textContent).toBe('Sam Hill');

    fireEvent.pointerOut(avatar, { relatedTarget: document.body });
    expect(screen.queryByRole('tooltip')).toBeNull();
  });

  it('swaps straight to the next avatar once one name has shown', () => {
    render(
      <>
        <AssigneeAvatar name="Sam Hill" initials="SH" seed="7" />
        <AssigneeAvatar name="Ria Moss" initials="RM" seed="8" />
        <AvatarNameTooltip />
      </>
    );
    const [sam, ria] = screen.getAllByRole('img');
    hover(sam);
    act(() => {
      vi.advanceTimersByTime(200);
    });
    fireEvent.pointerOut(sam, { relatedTarget: ria });
    hover(ria);
    expect(screen.getByRole('tooltip').textContent).toBe('Ria Moss');
  });

  it('stays closed for touch, so a tap never leaves it stuck open', () => {
    render(
      <>
        <Avatar initials="SH" seed="7" name="Sam Hill" />
        <AvatarNameTooltip />
      </>
    );
    hover(screen.getByRole('img', { name: 'Sam Hill' }), 'touch');
    act(() => {
      vi.advanceTimersByTime(500);
    });
    expect(screen.queryByRole('tooltip')).toBeNull();
  });

  it('closes on click and on Escape', () => {
    render(
      <>
        <button type="button" data-avatar-name="Sam Hill">
          <Avatar initials="SH" seed="7" aria-hidden />
        </button>
        <AvatarNameTooltip />
      </>
    );
    const button = screen.getByRole('button');
    hover(button);
    act(() => {
      vi.advanceTimersByTime(200);
    });
    expect(screen.getByRole('tooltip').textContent).toBe('Sam Hill');
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(screen.queryByRole('tooltip')).toBeNull();

    hover(document.body);
    hover(button);
    act(() => {
      vi.advanceTimersByTime(200);
    });
    fireEvent.pointerDown(button);
    expect(screen.queryByRole('tooltip')).toBeNull();
  });
});

describe('Avatar name', () => {
  it('labels the avatar with the name unless it is hidden', () => {
    render(
      <>
        <Avatar initials="SH" seed="7" name="Sam Hill" />
        <Avatar initials="RM" seed="8" name="Ria Moss" aria-hidden />
      </>
    );
    expect(screen.getByRole('img', { name: 'Sam Hill' })).toHaveAttribute(
      'data-avatar-name',
      'Sam Hill'
    );
    const hidden = document.querySelector('[data-avatar-name="Ria Moss"]');
    expect(hidden).not.toBeNull();
    expect(hidden).not.toHaveAttribute('role');
  });

  it('no longer leans on the slow native title tooltip', () => {
    render(<AssigneeAvatar name="Sam Hill" initials="SH" seed="7" />);
    const avatar = screen.getByRole('img', { name: 'Sam Hill' });
    expect(avatar).not.toHaveAttribute('title');
    expect(avatar).toHaveAttribute('data-avatar-name', 'Sam Hill');
  });
});
