import React from 'react';
import { act, fireEvent, render, screen } from '@testing-library/react';
import type { Editor } from '@tiptap/react';
import { describe, expect, it, vi } from 'vitest';
import DescriptionEditor from './DescriptionEditor';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

function renderEditor(value: string) {
  const onSave = vi.fn();
  render(<DescriptionEditor value={value} onSave={onSave} placeholder="Add a description" />);
  const box = screen.getByRole('textbox', { name: 'flow.taskDetail.descriptionLabel' });
  // Tiptap exposes the instance on its element; typing into a contenteditable
  // is not something the DOM shim can simulate.
  const editor = (box as HTMLElement & { editor: Editor }).editor;
  return { box, editor, onSave };
}

describe('DescriptionEditor', () => {
  it('keeps the line breaks and bullets of a plain-text description', () => {
    const { box } = renderEditor('USER STORY\nAs an owner\n\n- one\n- two');

    expect(box.querySelector('p br')).not.toBeNull();
    expect(box.querySelectorAll('ul > li')).toHaveLength(2);
  });

  it('does not rewrite a plain-text description that was clicked into but not edited', () => {
    // Placing the caret makes the editor add its trailing paragraph after the
    // list, which is a document change without being an edit.
    const { editor, onSave } = renderEditor('USER STORY\nAs an owner\n\n- one\n- two');

    act(() => {
      editor.commands.setTextSelection(2);
    });

    expect(onSave).not.toHaveBeenCalled();
  });

  it('saves an edit as Markdown', () => {
    const { editor, onSave } = renderEditor('Fix the flaky test');

    act(() => {
      editor.chain().selectAll().toggleBold().run();
    });

    expect(onSave).toHaveBeenCalledExactlyOnceWith('**Fix the flaky test**');
  });

  it('saves a checklist item ticked without focusing the editor', () => {
    const { box, onSave } = renderEditor('- [ ] Ship it');

    fireEvent.click(box.querySelector('input[type="checkbox"]') as HTMLInputElement);

    expect(onSave).toHaveBeenCalledExactlyOnceWith('- [x] Ship it');
  });
});
