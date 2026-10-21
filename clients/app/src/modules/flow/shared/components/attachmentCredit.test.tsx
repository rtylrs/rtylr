import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en' } }),
}));
vi.mock('@voxire/api', () => ({ uploadAsset: vi.fn() }));

const storeState = vi.hoisted(() => ({
  members: [] as { id: string; name: string }[],
}));
vi.mock('../stores/flowStore', () => ({
  useFlowStore: (selector: (s: typeof storeState) => unknown) => selector(storeState),
}));

import AttachmentUpload from './AttachmentUpload';

const attachment = {
  id: 'a1',
  name: 'contract.pdf',
  size: 48128,
  uploadedBy: '9403',
  url: 'https://example.test/contract.pdf',
  mime: 'application/pdf',
  createdAt: '2026-01-01T00:00:00Z',
};

describe('who an attachment is credited to', () => {
  beforeEach(() => {
    storeState.members = [];
  });

  it('names the person rather than printing their id', () => {
    // The row read "47.0 KB · 9403", crediting the uploader with a database key
    // (rtylr#1079).
    storeState.members = [{ id: '9403', name: 'Farah Osman' }];
    render(<AttachmentUpload attachments={[attachment] as never} />);

    expect(screen.getByText(/Farah Osman/)).toBeInTheDocument();
    expect(screen.queryByText(/9403/)).toBeNull();
  });

  it('matches on value, not on type', () => {
    // Task ids arrive as numbers in some payloads and strings in others, which
    // is the same int64-as-a-wire-string seam as rtylr#1226.
    storeState.members = [{ id: 9403 as unknown as string, name: 'Farah Osman' }];
    render(<AttachmentUpload attachments={[attachment] as never} />);

    expect(screen.getByText(/Farah Osman/)).toBeInTheDocument();
  });

  it('falls back to the id rather than to nothing when no member matches', () => {
    // An unresolvable id is still information. Rendering an empty string would
    // read as "uploaded by nobody".
    storeState.members = [{ id: '1', name: 'Someone Else' }];
    render(<AttachmentUpload attachments={[attachment] as never} />);

    expect(screen.getByText(/9403/)).toBeInTheDocument();
  });
});
