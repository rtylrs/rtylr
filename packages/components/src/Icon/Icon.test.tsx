import { render, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import Icon from './index';

describe('Icon component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders SVG content when fetch succeeds with valid SVG', async () => {
    const mockSvg = '<svg><path d="M0 0h24v24H0z"/></svg>';
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
      ok: true,
      status: 200,
      text: async () => mockSvg,
    } as Response);

    const { container } = render(<Icon icon="user" title="User Icon" />);

    await waitFor(() => {
      expect(container.querySelector('span')).not.toBeNull();
    });

    expect(container.querySelector('svg')).not.toBeNull();
  });

  it('sanitizes path traversal in icon prop', async () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
      ok: true,
      status: 200,
      text: async () => '<svg></svg>',
    } as Response);

    render(<Icon icon="../../../etc/passwd" />);

    await waitFor(() => {
      expect(fetchSpy).toHaveBeenCalledWith('/icons/etcpasswd.svg');
    });
  });

  it('rejects non-SVG content (e.g. 404 HTML response)', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
      ok: false,
      status: 404,
      text: async () => '<html>404 Not Found</html>',
    } as Response);

    const { container } = render(<Icon icon="non-existent" />);

    await waitFor(() => {
      expect(container.querySelector('span')).toBeNull();
    });
  });
});
