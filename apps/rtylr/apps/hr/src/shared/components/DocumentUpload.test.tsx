import React from 'react';
import { fireEvent, render, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  uploadAsset: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  uploadAsset: mocks.uploadAsset,
}));

import DocumentUpload from './DocumentUpload';

const pdf = (name: string): File => new File(['x'], name, { type: 'application/pdf' });

describe('DocumentUpload in-flight guard', () => {
  it('ignores a second selection while an upload is still running', async () => {
    let resolveUpload: (v: { url: string; size: number; mime: string }) => void = () => {};
    mocks.uploadAsset.mockReturnValue(
      new Promise((resolve) => {
        resolveUpload = resolve;
      })
    );
    const onFile = vi.fn();
    const { container } = render(<DocumentUpload onFile={onFile} />);
    const dropZone = container.firstElementChild as HTMLElement;
    const input = container.querySelector('input[type="file"]') as HTMLInputElement;

    fireEvent.drop(dropZone, { dataTransfer: { files: [pdf('a.pdf')] } });
    expect(mocks.uploadAsset).toHaveBeenCalledTimes(1);

    // Drop and file-input selections bypass the click handler's `uploading`
    // check, so both used to start a second concurrent upload.
    fireEvent.drop(dropZone, { dataTransfer: { files: [pdf('b.pdf')] } });
    fireEvent.change(input, { target: { files: [pdf('c.pdf')] } });

    expect(mocks.uploadAsset).toHaveBeenCalledTimes(1);
    expect(onFile).not.toHaveBeenCalled();

    resolveUpload({ url: 'https://cdn.test/a.pdf', size: 2048, mime: 'application/pdf' });

    await waitFor(() => expect(onFile).toHaveBeenCalledTimes(1));
    expect(onFile.mock.calls[0][0]).toMatchObject({ name: 'a.pdf', url: 'https://cdn.test/a.pdf' });

    // The guard releases once the request settles; a later file is accepted.
    mocks.uploadAsset.mockResolvedValue({
      url: 'https://cdn.test/b.pdf',
      size: 1024,
      mime: 'application/pdf',
    });
    fireEvent.drop(dropZone, { dataTransfer: { files: [pdf('b.pdf')] } });

    await waitFor(() => expect(onFile).toHaveBeenCalledTimes(2));
    expect(onFile.mock.calls[1][0]).toMatchObject({ name: 'b.pdf', url: 'https://cdn.test/b.pdf' });
  });
});
