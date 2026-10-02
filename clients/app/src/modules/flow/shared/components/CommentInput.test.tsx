import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { CommentInput } from './CommentInput';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

describe('CommentInput', () => {
  it('shows a translated reason on the disabled emoji picker button instead of a silent disabled control', () => {
    render(<CommentInput onSubmit={() => {}} />);

    const emojiButton = screen.getByRole('button', { name: 'flowCommon.emojiPickerComingSoon' });

    expect(emojiButton).toBeDisabled();
  });
});
