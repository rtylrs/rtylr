import { describe, it, expect, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import {
  CommentInput,
  findActiveMention,
  matchMentionCandidates,
  type MentionCandidate,
} from './CommentInput';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const people: MentionCandidate[] = [
  { id: '7', name: 'Mohamad Issa', initials: 'MI' },
  { id: '8', name: 'Abed Karam', initials: 'AK' },
  { id: '9', name: 'Maya Haddad', initials: 'MH' },
];

function type(textarea: HTMLElement, value: string) {
  fireEvent.change(textarea, { target: { value, selectionStart: value.length } });
}

describe('CommentInput', () => {
  it('shows a translated reason on the disabled emoji picker button instead of a silent disabled control', () => {
    render(<CommentInput onSubmit={() => {}} />);

    const emojiButton = screen.getByRole('button', { name: 'flowCommon.emojiPickerComingSoon' });

    expect(emojiButton).toBeDisabled();
  });

  it('lists the people matching what follows the @ and inserts the picked full name', () => {
    const onSubmit = vi.fn();
    render(<CommentInput onSubmit={onSubmit} mentionCandidates={people} />);
    const textarea = screen.getByRole('combobox');

    type(textarea, 'Ready for @iss');
    const options = screen.getAllByRole('option');
    expect(options.map((option) => option.textContent)).toEqual(['MIMohamad Issa']);

    fireEvent.keyDown(textarea, { key: 'Enter' });
    expect(textarea).toHaveValue('Ready for @Mohamad Issa ');
    expect(screen.queryByRole('listbox')).toBeNull();
    // Enter picked the person; it did not post the comment.
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it('says there is no one to mention instead of opening an empty list', () => {
    // The picker used to open as a blank strip when the member list was empty.
    render(<CommentInput onSubmit={() => {}} mentionCandidates={[]} />);

    type(screen.getByRole('combobox'), '@');

    expect(screen.getByRole('listbox')).toHaveTextContent('flow.taskDetail.noOneToMention');
  });

  it('opens the picker from the @ button', () => {
    render(<CommentInput onSubmit={() => {}} mentionCandidates={people} />);

    fireEvent.click(screen.getByRole('button', { name: 'flow.taskDetail.mentionSomeone' }));

    expect(screen.getByRole('combobox')).toHaveValue('@');
    expect(screen.getAllByRole('option')).toHaveLength(3);
  });
});

describe('findActiveMention', () => {
  it('finds the mention being typed at the caret', () => {
    expect(findActiveMention('hi @moh', 7)).toEqual({ start: 3, query: 'moh' });
    expect(findActiveMention('@Mohamad Is', 11)).toEqual({ start: 0, query: 'Mohamad Is' });
  });

  it('ignores an @ inside a word, such as an email address', () => {
    expect(findActiveMention('mail sam@site.io', 16)).toBeNull();
  });

  it('ignores an @ on an earlier line', () => {
    expect(findActiveMention('@moh\nnext line', 14)).toBeNull();
  });
});

describe('matchMentionCandidates', () => {
  it('matches the start of any word of the name, ignoring case', () => {
    expect(matchMentionCandidates(people, 'm').map((p) => p.id)).toEqual(['7', '9']);
    expect(matchMentionCandidates(people, 'HADD').map((p) => p.id)).toEqual(['9']);
  });
});
