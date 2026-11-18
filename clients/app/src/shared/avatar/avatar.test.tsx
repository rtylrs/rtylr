import React from 'react';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { Avatar, avatarColorAt } from '@voxire/ui';
import { AVATAR_STYLES, randomAvatarSeed } from './avatarModel';
import { renderIllustration } from './illustrated';
import { toChosenAvatar } from './useAvatars';

const api = vi.hoisted(() => ({
  updateAvatar: vi.fn(),
  listBusinessAvatars: vi.fn(),
  uploadAsset: vi.fn(),
}));
vi.mock('../lib/clients', () => ({
  userClient: { updateAvatar: api.updateAvatar, listBusinessAvatars: api.listBusinessAvatars },
}));
vi.mock('@voxire/api', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@voxire/api')>()),
  uploadAsset: api.uploadAsset,
}));
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, values?: Record<string, unknown>) =>
      values ? `${key} ${JSON.stringify(values)}` : key,
  }),
}));

import { toast } from 'sonner';
import { useUserStore } from '../stores/userStore';
import { AvatarDirectoryProvider } from './AvatarDirectoryProvider';
import { AvatarPicker, initialsOf } from './AvatarPicker';

const withQuery = (ui: React.ReactElement) => (
  <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
    {ui}
  </QueryClientProvider>
);

beforeEach(() => {
  vi.clearAllMocks();
  api.listBusinessAvatars.mockResolvedValue({ avatars: [] });
  useUserStore.setState({ user: { id: 7, businessId: 1, fullName: 'Abed Amouneh' } });
});

describe('avatar model', () => {
  it('offers the twenty chosen styles, credits every CC BY one, and has unique ids', () => {
    expect(AVATAR_STYLES).toHaveLength(20);
    expect(new Set(AVATAR_STYLES.map((s) => s.id)).size).toBe(20);
    const credited = AVATAR_STYLES.filter((s) => s.credit?.license === 'CC BY 4.0').map(
      (s) => s.id
    );
    expect(credited.sort()).toEqual(['adventurer', 'big-smile', 'croodles', 'micah', 'personas']);
  });

  it('makes seeds the server accepts', () => {
    for (let i = 0; i < 50; i += 1) expect(randomAvatarSeed()).toMatch(/^[A-Za-z0-9_-]{1,64}$/);
  });

  it('reads only kinds the portal can draw from the wire', () => {
    expect(toChosenAvatar({ kind: 'emoji', emoji: '🗿', color: 2 })).toEqual({
      kind: 'emoji',
      emoji: '🗿',
      color: 2,
      url: undefined,
      style: undefined,
      seed: undefined,
    });
    expect(toChosenAvatar({ kind: 'hologram' })).toBeUndefined();
    expect(toChosenAvatar(undefined)).toBeUndefined();
  });

  it('takes initials from the first and last name', () => {
    expect(initialsOf('Abed El-Fattah Amouneh')).toBe('AA');
    expect(initialsOf('Issa')).toBe('I');
    expect(initialsOf('')).toBe('?');
  });
});

describe('illustrated avatars', () => {
  it('renders the same picture for the same seed, as an SVG data URI', async () => {
    const a = await renderIllustration('gaze', 'k3x9a');
    const b = await renderIllustration('gaze', 'k3x9a');
    expect(a).toMatch(/^data:image\/svg\+xml/);
    expect(a).toBe(b);
    expect(await renderIllustration('gaze', 'other')).not.toBe(a);
  });

  it('refuses a style it does not ship', async () => {
    await expect(renderIllustration('pepe', 'x')).rejects.toThrow();
  });
});

describe('avatar directory', () => {
  it("shows a colleague's chosen emoji and colour wherever their id is the seed", async () => {
    api.listBusinessAvatars.mockResolvedValue({
      avatars: [
        { userId: 12, avatar: { kind: 'emoji', emoji: '🦊', color: 4 } },
        { userId: 13, avatar: { kind: 'initials', color: 9 } },
      ],
    });
    render(
      withQuery(
        <AvatarDirectoryProvider>
          <Avatar seed="12" initials="AH" />
          <Avatar seed="13" initials="MH" />
          <Avatar seed="14" initials="MI" />
        </AvatarDirectoryProvider>
      )
    );
    expect(await screen.findByText('🦊')).toBeTruthy();
    const mh = screen.getByText('MH').parentElement as HTMLElement;
    expect(mh.style.background.toLowerCase()).toBe(avatarColorAt(9).background.toLowerCase());
    // Someone on the default still shows initials.
    expect(screen.getByText('MI')).toBeTruthy();
  });

  it("uses the signed-in user's own choice straight from their profile", async () => {
    useUserStore.setState({
      user: {
        id: 7,
        businessId: 1,
        fullName: 'Abed',
        avatar: { kind: 'emoji', emoji: '🗿', color: 1 },
      },
    });
    render(
      withQuery(
        <AvatarDirectoryProvider>
          <Avatar seed="7" initials="AA" />
        </AvatarDirectoryProvider>
      )
    );
    expect(await screen.findByText('🗿')).toBeTruthy();
  });
});

describe('avatar picker', () => {
  const open = () =>
    render(
      withQuery(
        <AvatarPicker
          open
          onOpenChange={vi.fn()}
          userId="7"
          name="Abed Amouneh"
          current={undefined}
        />
      )
    );

  it('saves a picked emoji with its colour', async () => {
    api.updateAvatar.mockResolvedValue({
      user: { avatar: { kind: 'emoji', emoji: '💀', color: 3 } },
    });
    open();
    fireEvent.click(screen.getByRole('tab', { name: 'settings.avatar.tabs.emoji' }));
    fireEvent.click(screen.getByRole('option', { name: '💀' }));
    fireEvent.click(screen.getByRole('radio', { name: 'settings.avatar.colourN {"n":4}' }));
    fireEvent.click(screen.getByRole('button', { name: 'settings.avatar.save' }));
    await waitFor(() =>
      expect(api.updateAvatar).toHaveBeenCalledWith({
        avatar: { kind: 'emoji', emoji: '💀', color: 3 },
      })
    );
    await waitFor(() => expect(useUserStore.getState().user?.avatar?.emoji).toBe('💀'));
  });

  it('saves an illustration from the grid, and shows the credit for a CC BY style', async () => {
    api.updateAvatar.mockResolvedValue({ user: {} });
    open();
    fireEvent.click(screen.getByRole('button', { name: /^Micah/ }));
    expect(document.querySelector('[data-avatar-credit]')?.textContent).toContain('Micah Lanier');
    const [first] = screen.getAllByRole('option');
    fireEvent.click(first);
    fireEvent.click(screen.getByRole('button', { name: 'settings.avatar.save' }));
    await waitFor(() => expect(api.updateAvatar).toHaveBeenCalled());
    const sent = api.updateAvatar.mock.calls[0][0].avatar;
    expect(sent.kind).toBe('illustrated');
    expect(sent.style).toBe('micah');
    expect(sent.seed).toMatch(/^[A-Za-z0-9_-]+$/);
  });

  it('refuses a photo over 5 MB before uploading anything', () => {
    open();
    fireEvent.click(screen.getByRole('tab', { name: 'settings.avatar.tabs.photo' }));
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    const big = new File([new Uint8Array(5 * 1024 * 1024 + 1)], 'big.png', { type: 'image/png' });
    act(() => {
      fireEvent.change(input, { target: { files: [big] } });
    });
    expect(toast.error).toHaveBeenCalledWith('settings.avatar.tooLarge');
    expect(api.uploadAsset).not.toHaveBeenCalled();
  });

  it('goes back to the default', async () => {
    api.updateAvatar.mockResolvedValue({ user: {} });
    render(
      withQuery(
        <AvatarPicker
          open
          onOpenChange={vi.fn()}
          userId="7"
          name="Abed"
          current={{ kind: 'emoji', emoji: '🗿', color: 1 }}
        />
      )
    );
    fireEvent.click(screen.getByRole('button', { name: 'settings.avatar.useDefault' }));
    await waitFor(() => expect(api.updateAvatar).toHaveBeenCalledWith({}));
  });
});
