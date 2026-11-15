import React, { useCallback, useEffect, useRef } from 'react';
import { EditorContent, useEditor, useEditorState, type Editor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import { TaskItem, TaskList } from '@tiptap/extension-list';
import { Placeholder } from '@tiptap/extensions';
import { Markdown } from '@tiptap/markdown';
import {
  Bold,
  Code,
  Heading2,
  Italic,
  List,
  ListChecks,
  ListOrdered,
  Quote,
  Strikethrough,
  type LucideIcon,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';

interface DescriptionEditorProps {
  /** Markdown. Plain-text descriptions written before the editor are valid input. */
  value: string;
  onSave: (markdown: string) => void;
  placeholder: string;
  disabled?: boolean;
}

interface FormatAction {
  key: string;
  icon: LucideIcon;
  isActive: (editor: Editor) => boolean;
  run: (editor: Editor) => void;
}

const FORMAT_ACTIONS: FormatAction[] = [
  {
    key: 'bold',
    icon: Bold,
    isActive: (e) => e.isActive('bold'),
    run: (e) => e.chain().focus().toggleBold().run(),
  },
  {
    key: 'italic',
    icon: Italic,
    isActive: (e) => e.isActive('italic'),
    run: (e) => e.chain().focus().toggleItalic().run(),
  },
  {
    key: 'strikethrough',
    icon: Strikethrough,
    isActive: (e) => e.isActive('strike'),
    run: (e) => e.chain().focus().toggleStrike().run(),
  },
  {
    key: 'code',
    icon: Code,
    isActive: (e) => e.isActive('code'),
    run: (e) => e.chain().focus().toggleCode().run(),
  },
  {
    key: 'heading',
    icon: Heading2,
    isActive: (e) => e.isActive('heading', { level: 2 }),
    run: (e) => e.chain().focus().toggleHeading({ level: 2 }).run(),
  },
  {
    key: 'bulletList',
    icon: List,
    isActive: (e) => e.isActive('bulletList'),
    run: (e) => e.chain().focus().toggleBulletList().run(),
  },
  {
    key: 'numberedList',
    icon: ListOrdered,
    isActive: (e) => e.isActive('orderedList'),
    run: (e) => e.chain().focus().toggleOrderedList().run(),
  },
  {
    key: 'checklist',
    icon: ListChecks,
    isActive: (e) => e.isActive('taskList'),
    run: (e) => e.chain().focus().toggleTaskList().run(),
  },
  {
    key: 'quote',
    icon: Quote,
    isActive: (e) => e.isActive('blockquote'),
    run: (e) => e.chain().focus().toggleBlockquote().run(),
  },
];

// A list is followed by an empty paragraph to click into; it is not content.
function toMarkdown(editor: Editor): string {
  return editor.getMarkdown().trimEnd();
}

const DescriptionEditor: React.FC<DescriptionEditorProps> = ({
  value,
  onSave,
  placeholder,
  disabled = false,
}) => {
  const { t } = useTranslation();
  const onSaveRef = useRef(onSave);
  useEffect(() => {
    onSaveRef.current = onSave;
  }, [onSave]);

  // `loaded` is the value last taken from or sent to the server. `baseline` is
  // the editor's own Markdown for it: a plain-text description serializes
  // differently from how it was typed, and that difference is not an edit.
  const loaded = useRef(value);
  const baseline = useRef<string | null>(null);
  const pending = useRef<string | null>(null);
  const flush = useCallback(() => {
    const markdown = pending.current;
    pending.current = null;
    if (markdown === null || markdown === baseline.current) return;
    baseline.current = markdown;
    loaded.current = markdown;
    onSaveRef.current(markdown);
  }, []);

  const editor = useEditor({
    extensions: [
      // Markdown has no underline, so it could not be stored.
      StarterKit.configure({ underline: false, heading: { levels: [1, 2, 3] } }),
      TaskList,
      TaskItem.configure({ nested: true }),
      Placeholder.configure({ placeholder }),
      // `breaks` keeps the single line breaks of plain-text descriptions.
      Markdown.configure({ markedOptions: { gfm: true, breaks: true } }),
    ],
    content: value,
    contentType: 'markdown',
    editable: !disabled,
    editorProps: {
      attributes: {
        class: 'flow-description-content',
        role: 'textbox',
        'aria-multiline': 'true',
        'aria-label': t('flow.taskDetail.descriptionLabel'),
        dir: 'auto',
      },
    },
    onUpdate: ({ editor: current }) => {
      pending.current = toMarkdown(current);
      // Ticking a checklist item changes the text without focusing the editor,
      // so no blur will follow to save it.
      if (!current.isFocused) flush();
    },
    onBlur: flush,
  });

  useEffect(() => flush, [flush]);

  // Tiptap reports its own creation a tick late, after a click could land.
  useEffect(() => {
    if (editor && baseline.current === null) baseline.current = toMarkdown(editor);
  }, [editor]);

  useEffect(() => {
    editor?.setEditable(!disabled, false);
  }, [editor, disabled]);

  useEffect(() => {
    if (!editor || editor.isFocused || value === loaded.current) return;
    loaded.current = value;
    pending.current = null;
    editor.commands.setContent(value, { contentType: 'markdown', emitUpdate: false });
    baseline.current = toMarkdown(editor);
  }, [editor, value]);

  const active = useEditorState({
    editor,
    selector: ({ editor: current }) =>
      FORMAT_ACTIONS.map((action) => (current ? action.isActive(current) : false)),
  });

  return (
    <div className="flow-description" data-editable={!disabled}>
      <EditorContent editor={editor} />
      {editor && !disabled && (
        <div
          role="toolbar"
          aria-label={t('flow.taskDetail.descriptionFormat.toolbar')}
          className="flow-description-toolbar"
        >
          {FORMAT_ACTIONS.map((action, index) => {
            const label = t(`flow.taskDetail.descriptionFormat.${action.key}`);
            return (
              <button
                key={action.key}
                type="button"
                title={label}
                aria-label={label}
                aria-pressed={active?.[index] ?? false}
                // Keep the caret in the editor while the button is clicked.
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => action.run(editor)}
              >
                <action.icon size={14} aria-hidden />
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default DescriptionEditor;
