// @vitest-environment happy-dom
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeAll, describe, expect, it, vi } from 'vitest';
import Settings from './Settings';
import { initTestI18n } from '../../shared/testing/i18n';

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));
vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));
vi.mock('../../shared/components/PageHeader', () => ({
  default: ({ title }: { title: string }) => <div>{title}</div>,
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Finance team settings', () => {
  it('links team access management to Dash', () => {
    render(<Settings />);
    fireEvent.click(screen.getByRole('button', { name: 'Team' }));

    expect(screen.getByRole('link', { name: 'Open team management' })).toHaveAttribute(
      'href',
      expect.stringMatching(/\/team$/)
    );
  });
});
