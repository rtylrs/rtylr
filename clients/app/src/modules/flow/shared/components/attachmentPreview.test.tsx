import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactElement } from 'react';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en' } }),
}));
const signAssetUrl = vi.hoisted(() => vi.fn());
vi.mock('@voxire/api', () => ({ uploadAsset: vi.fn(), signAssetUrl }));
vi.mock('@voxire/config', () => ({ getCdnBaseUrl: () => 'https://api.voxire.com' }));
vi.mock('../stores/flowStore', () => ({
  useFlowStore: (selector: (s: { members: [] }) => unknown) => selector({ members: [] }),
}));

import AttachmentUpload from './AttachmentUpload';

const CDN_FILE = 'https://api.voxire.com/files/0f1e2d3c-4b5a-4678-9abc-def012345678';
const SIGNED = `${CDN_FILE}?expires=1&sig=abc`;

const image = {
  id: 'a1',
  name: 'statement-en-phone.png',
  size: 641741,
  uploadedBy: '18',
  url: CDN_FILE,
  mime: 'image/*',
  createdAt: '2026-11-15T00:00:00Z',
};

function renderWithQuery(ui: ReactElement) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const view = render(<QueryClientProvider client={client}>{ui}</QueryClientProvider>);
  return {
    ...view,
    rerender: (next: ReactElement) =>
      view.rerender(<QueryClientProvider client={client}>{next}</QueryClientProvider>),
  };
}

describe('image attachments', () => {
  beforeEach(() => {
    signAssetUrl.mockReset();
    signAssetUrl.mockResolvedValue({ url: SIGNED, expiresAt: 1 });
  });

  it('loads a CDN upload through a signed link and opens it full size', async () => {
    // Uploads are private, so the stored address alone answers 404 to an
    // <img>, and every thumbnail fell back to the placeholder icon.
    renderWithQuery(<AttachmentUpload attachments={[image] as never} />);

    const thumb = await screen.findByRole('img', { name: image.name });
    expect(thumb).toHaveAttribute('src', SIGNED);
    // The thumbnail is a resized copy, signed by path so a file stored under
    // an earlier CDN host still signs.
    const id = new URL(CDN_FILE).pathname.split('/').pop();
    expect(signAssetUrl).toHaveBeenCalledWith(
      `/images/${id}`,
      expect.objectContaining({ params: { w: '160' } })
    );

    fireEvent.click(screen.getByRole('button', { name: 'flow.taskDetail.openAttachment' }));

    const dialog = await screen.findByRole('dialog');
    await waitFor(() => expect(dialog.querySelector(`img[src="${SIGNED}"]`)).not.toBeNull());
  });

  it('uses an external link as stored, without signing it', async () => {
    const external = { ...image, url: 'https://example.test/shot.png' };
    renderWithQuery(<AttachmentUpload attachments={[external] as never} />);

    expect(await screen.findByRole('img', { name: image.name })).toHaveAttribute(
      'src',
      external.url
    );
    expect(signAssetUrl).not.toHaveBeenCalled();
  });

  it('falls back to the stored address when signing fails', async () => {
    signAssetUrl.mockRejectedValue(new Error('signing is not configured'));
    renderWithQuery(<AttachmentUpload attachments={[image] as never} />);

    // The hook retries signing once before falling back.
    expect(await screen.findByRole('img', { name: image.name }, { timeout: 4000 })).toHaveAttribute(
      'src',
      CDN_FILE
    );
  });

  it('does not load or link an address that is not http(s)', () => {
    renderWithQuery(
      <AttachmentUpload attachments={[{ ...image, url: 'javascript:alert(1)' }] as never} />
    );

    expect(screen.queryByRole('img', { name: image.name })).toBeNull();
    expect(screen.getByRole('button', { name: image.name })).toBeDisabled();
  });
});

describe('removing attachments', () => {
  it('offers remove only on attachments the viewer may delete', () => {
    // The server lets the uploader remove their own; the button used to show
    // on every attachment and fail with "Failed to remove attachment".
    signAssetUrl.mockResolvedValue({ url: SIGNED, expiresAt: 1 });
    renderWithQuery(
      <AttachmentUpload
        attachments={[image, { ...image, id: 'a2', uploadedBy: '20' }] as never}
        onRemove={() => {}}
        canRemove={(attachment) => attachment.uploadedBy === '18'}
      />
    );

    expect(
      screen.getAllByRole('button', { name: 'flow.taskDetail.removeAttachmentLabel' })
    ).toHaveLength(1);
  });
});

describe('thumbnail fallback', () => {
  beforeEach(() => {
    signAssetUrl.mockReset();
    signAssetUrl.mockImplementation(async (path: string) => ({
      url: `https://api.voxire.com${path}?sig=s`,
      expiresAt: Math.floor(Date.now() / 1000) + 3600,
    }));
  });

  it('loads the original when the resized copy fails', async () => {
    renderWithQuery(<AttachmentUpload attachments={[image] as never} />);
    const thumb = await screen.findByRole('img', { name: image.name });
    expect(thumb.getAttribute('src')).toContain('/images/');

    fireEvent.error(thumb);

    await waitFor(() =>
      expect(screen.getByRole('img', { name: image.name }).getAttribute('src')).toContain('/files/')
    );
  });

  it('tries the resized copy again when a row shows a different file', async () => {
    // The fallback state used to survive a url change, so a reused row skipped
    // the next file's resized copy and loaded its full-size original.
    const view = renderWithQuery(<AttachmentUpload attachments={[image] as never} />);
    fireEvent.error(await screen.findByRole('img', { name: image.name }));
    await waitFor(() =>
      expect(screen.getByRole('img', { name: image.name }).getAttribute('src')).toContain('/files/')
    );

    const next = { ...image, url: 'https://api.voxire.com/files/11111111-2222-4333-8444-555555555555' };
    view.rerender(<AttachmentUpload attachments={[next] as never} />);

    await waitFor(() =>
      expect(screen.getByRole('img', { name: image.name }).getAttribute('src')).toContain(
        '/images/11111111-2222-4333-8444-555555555555'
      )
    );
  });
});
