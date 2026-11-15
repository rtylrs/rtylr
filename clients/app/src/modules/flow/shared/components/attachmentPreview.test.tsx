import { describe, it, expect, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en' } }),
}));
vi.mock('@voxire/api', () => ({ uploadAsset: vi.fn() }));
vi.mock('../stores/flowStore', () => ({
  useFlowStore: (selector: (s: { members: [] }) => unknown) => selector({ members: [] }),
}));

import AttachmentUpload from './AttachmentUpload';

const image = {
  id: 'a1',
  name: 'statement-en-phone.png',
  size: 641741,
  uploadedBy: '18',
  url: 'https://cdn.example.test/statement-en-phone.png',
  mime: 'image/*',
  createdAt: '2026-11-15T00:00:00Z',
};

describe('image attachments', () => {
  it('shows the image and opens it full size', () => {
    render(<AttachmentUpload attachments={[image] as never} />);

    const thumb = screen.getByRole('img', { name: image.name });
    expect(thumb).toHaveAttribute('src', image.url);

    fireEvent.click(screen.getByRole('button', { name: 'flow.taskDetail.openAttachment' }));

    const dialog = screen.getByRole('dialog');
    expect(dialog.querySelector(`img[src="${image.url}"]`)).not.toBeNull();
  });

  it('does not load or link an address that is not http(s)', () => {
    render(<AttachmentUpload attachments={[{ ...image, url: 'javascript:alert(1)' }] as never} />);

    expect(screen.queryByRole('img', { name: image.name })).toBeNull();
    expect(screen.getByRole('button', { name: image.name })).toBeDisabled();
  });
});
