// @vitest-environment happy-dom
import React from 'react';
import { render, waitFor, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { uploadAssetMock } = vi.hoisted(() => ({ uploadAssetMock: vi.fn() }));

vi.mock('@voxire/api', () => ({
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
  uploadAsset: uploadAssetMock,
}));

vi.mock('@voxire/config', () => ({ cdnImageUrl: (url: string) => url }));

vi.mock('react-i18next', () => ({ useTranslation: () => ({ t: (key: string) => key }) }));

import ReceiptUpload from './ReceiptUpload';

const imageFile = (name: string): File => new File(['x'], name, { type: 'image/png' });

describe('ReceiptUpload', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('ignores a drop that arrives while an upload is still in flight', async () => {
    let resolveUpload: (value: { url: string }) => void = () => {};
    uploadAssetMock.mockReturnValueOnce(
      new Promise<{ url: string }>((resolve) => {
        resolveUpload = resolve;
      })
    );

    const onChange = vi.fn();
    const { container } = render(<ReceiptUpload value={null} onChange={onChange} />);
    const dropZone = container.firstElementChild as HTMLElement;

    fireEvent.drop(dropZone, { dataTransfer: { files: [imageFile('first.png')] } });
    fireEvent.drop(dropZone, { dataTransfer: { files: [imageFile('second.png')] } });

    // The second drop must not start a competing upload whose result could
    // overwrite the first.
    expect(uploadAssetMock).toHaveBeenCalledTimes(1);
    expect(uploadAssetMock.mock.calls[0]?.[0]).toMatchObject({ name: 'first.png' });

    resolveUpload({ url: '/cdn/first.png' });
    await waitFor(() => expect(onChange).toHaveBeenCalledWith('/cdn/first.png'));
    expect(onChange).toHaveBeenCalledTimes(1);
  });
});
