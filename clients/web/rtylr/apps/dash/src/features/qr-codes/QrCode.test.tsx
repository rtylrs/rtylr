import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

const userState = vi.hoisted(() => ({
  user: { businessId: 1, businessSlug: 'acme' } as {
    logo?: string;
    businessId?: number;
    businessSlug?: string;
  } | null,
}));

vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: () => ({ user: userState.user }),
}));

const qrState = vi.hoisted(() => ({
  renders: [] as Array<Record<string, unknown>>,
}));

vi.mock('react-qrcode-logo', () => ({
  QRCode: (props: Record<string, unknown>) => {
    qrState.renders.push(props);
    return null;
  },
}));

const saveAsMock = vi.hoisted(() => vi.fn());
vi.mock('file-saver', () => ({ saveAs: saveAsMock }));

const html2canvasMock = vi.hoisted(() => vi.fn());
vi.mock('html2canvas', () => ({ default: html2canvasMock }));

import { toast } from 'sonner';
import QrCode from './QrCode';

const mainQrProps = (): Record<string, unknown> | undefined =>
  qrState.renders.filter((props) => props.size === 250).at(-1);

describe('QrCode generator', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    qrState.renders = [];
    userState.user = { businessId: 1, businessSlug: 'acme' };
  });

  it('stores number inputs as numbers so the eye radius takes effect', () => {
    const { container } = render(<QrCode />);

    const eyeRadius = container.querySelector('input[name="eyeRadius"]') as HTMLInputElement;
    fireEvent.change(eyeRadius, { target: { value: '20' } });

    const props = mainQrProps();
    expect(props?.eyeRadius).toBe(20);
    expect(typeof props?.eyeRadius).toBe('number');

    // Clearing the field must not leave NaN or a string behind either.
    fireEvent.change(eyeRadius, { target: { value: '' } });
    const cleared = mainQrProps()?.eyeRadius;
    expect(typeof cleared).toBe('number');
    expect(Number.isNaN(cleared)).toBe(false);
  });

  it('allows downloading with Display Logo checked when the account has no logo', async () => {
    userState.user = { businessId: 1 };
    html2canvasMock.mockResolvedValue({
      toBlob: (callback: (blob: Blob) => void) => callback(new Blob()),
    });

    render(<QrCode />);

    fireEvent.click(screen.getByRole('checkbox'));
    fireEvent.click(screen.getByRole('button', { name: 'Download QR Code' }));

    await waitFor(() => expect(saveAsMock).toHaveBeenCalledWith(expect.any(Blob), 'qr-code.png'));
    expect(toast.error).not.toHaveBeenCalled();
  });
});
