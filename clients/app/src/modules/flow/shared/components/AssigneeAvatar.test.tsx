import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { Avatar, avatarColor } from '@voxire/ui';
import { AssigneeAvatar, AvatarStack } from './AssigneeAvatar';

// happy-dom keeps a colour as written, a browser normalises it to rgb(); compare
// the hex form case-insensitively so the test reads the value that was set.
const rgb = (hex: string): string => hex.toLowerCase();
const bg = (el: Element): string => (el as HTMLElement).style.background.toLowerCase();

describe('avatar colours', () => {
  it('paints two people with the same initials in their own colours', () => {
    // AH could be Ahmad Homsi or Ahmed Hamad: the colour comes from the id.
    const a = avatarColor('12');
    const b = avatarColor('19');
    expect(a.background).not.toBe(b.background);
    render(
      <>
        <AssigneeAvatar name="Ahmad Homsi" initials="AH" seed="12" />
        <AssigneeAvatar name="Ahmed Hamad" initials="AH" seed="19" />
      </>
    );
    const [first, second] = screen.getAllByText('AH');
    expect(bg(first)).toBe(rgb(a.background));
    expect(bg(second)).toBe(rgb(b.background));
  });

  it('keeps an explicit colour when a caller passes one', () => {
    render(<AssigneeAvatar name="Team" initials="TM" color="#ff0000" seed="12" />);
    expect(screen.getByText('TM').style.color.toLowerCase()).toBe('#ff0000');
  });

  it('passes each member id through the stack', () => {
    render(
      <AvatarStack
        members={[
          { name: 'Abed', initials: 'AA', seed: '1' },
          { name: 'Ahmad', initials: 'AT', seed: '2' },
        ]}
      />
    );
    expect(bg(screen.getByText('AA'))).toBe(rgb(avatarColor('1').background));
  });

  it('colours the shared Avatar from its seed, and leaves a photo alone', () => {
    const { container } = render(
      <>
        <Avatar initials="MH" seed="7" />
        <Avatar initials="MI" seed="8" src="https://cdn.example/p.png" />
      </>
    );
    const [seeded, photo] = container.querySelectorAll('div');
    expect(bg(seeded)).toBe(rgb(avatarColor('7').background));
    expect(photo.hasAttribute('data-avatar-seeded')).toBe(false);
  });
});
