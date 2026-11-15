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

  // What the server is known to hold, and an edit not yet sent to it.
  const saved = useRef(value);
  const pending = useRef<string | null>(null);
  const flush = useCallback(() => {
    const markdown = pending.current;
    pending.current = null;
    if (markdown === null || markdown === saved.current) return;
    saved.current = markdown;
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
      // A list is followed by an empty paragraph to click into; it is not content.
      pending.current = current.getMarkdown().trimEnd();
      // Ticking a checklist item changes the text without focusing the editor,
      // so no blur will follow to save it.
      if (!current.isFocused) flush();
    },
    onBlur: flush,
  });

  useEffect(() => flush, [flush]);

  useEffect(() => {
    editor?.setEditable(!disabled, false);
  }, [editor, disabled]);

  useEffect(() => {
    if (!editor || editor.isFocused || value === saved.current) return;
    saved.current = value;
    pending.current = null;
    editor.commands.setContent(value, { contentType: 'markdown', emitUpdate: false });
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
