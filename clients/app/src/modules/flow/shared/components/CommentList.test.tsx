import { describe, it, expect, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import CommentList from './CommentList';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const comment = (id: string, authorId: string, body: string, edited = false) => ({
  id,
  taskId: 't1',
  authorId,
  author: authorId === '18' ? 'Ahmad Homsi' : 'Abed Karam',
  authorInitials: 'AH',
  body,
  createdAt: new Date().toISOString(),
  edited,
});

describe('CommentList', () => {
  it('offers Edit only on the signed-in user’s own comments and saves the new wording', () => {
    const onEdit = vi.fn();
    render(
      <CommentList
        comments={[comment('c1', '18', 'first wording'), comment('c2', '20', 'not mine')]}
        currentUserId="18"
        onEdit={onEdit}
      />
    );

    const editButtons = screen.getAllByRole('button', { name: 'flow.taskDetail.editCommentLabel' });
    expect(editButtons).toHaveLength(1);

    fireEvent.click(editButtons[0]);
    const box = screen.getByRole('combobox');
    expect(box).toHaveValue('first wording');
    fireEvent.change(box, { target: { value: 'second wording' } });
    fireEvent.click(screen.getByRole('button', { name: 'flow.taskDetail.saveComment' }));

    expect(onEdit).toHaveBeenCalledWith('c1', 'second wording');
    expect(screen.queryByRole('combobox')).toBeNull();
  });

  it('marks an edited comment', () => {
    render(<CommentList comments={[comment('c1', '18', 'reworded', true)]} />);

    expect(screen.getByText('flow.taskDetail.edited')).toBeInTheDocument();
  });

  it('highlights a full-name mention as one mention', () => {
    render(
      <CommentList
        comments={[comment('c1', '20', '@Mohamad Issa ready for your check')]}
        mentionCandidates={[{ id: '7', name: 'Mohamad Issa', initials: 'MI' }]}
      />
    );

    expect(screen.getByText('@Mohamad Issa')).toBeInTheDocument();
  });
});
