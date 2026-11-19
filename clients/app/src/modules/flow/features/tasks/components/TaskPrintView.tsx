import React, { useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';
import { useTranslation } from 'react-i18next';
import type { Comment } from '../../../shared/stores/flowStore';
import DescriptionEditor from './DescriptionEditor';

export interface PrintField {
  label: string;
  value: string;
}

interface TaskPrintViewProps {
  identifier: string;
  projectName: string;
  title: string;
  /** Markdown, as stored. */
  description: string;
  fields: PrintField[];
  subtasks: { id: string; title: string; completed: boolean }[];
  comments: Comment[];
  /** Called once the browser's print dialog has closed. */
  onDone: () => void;
}

// Tiptap paints the description a frame or two after mount. Printing before
// that leaves the description out of the PDF.
const PRINT_DELAY_MS = 150;
const PRINTING_CLASS = 'flow-printing';

/**
 * One task laid out for paper, printed with the browser's own dialog, which
 * is also how it becomes a PDF ("Save as PDF"). Nothing is rendered on a
 * server: the page is portalled to the body, and while it is printing the
 * print stylesheet hides everything else (see `.flow-task-print` in
 * globals.css). On screen it is never visible.
 */
const TaskPrintView: React.FC<TaskPrintViewProps> = ({
  identifier,
  projectName,
  title,
  description,
  fields,
  subtasks,
  comments,
  onDone,
}) => {
  const { t, i18n } = useTranslation();
  const onDoneRef = useRef(onDone);
  onDoneRef.current = onDone;

  useEffect(() => {
    const root = document.documentElement;
    root.classList.add(PRINTING_CLASS);
    const finish = (): void => onDoneRef.current();
    window.addEventListener('afterprint', finish);
    const timer = window.setTimeout(() => window.print(), PRINT_DELAY_MS);
    return () => {
      window.clearTimeout(timer);
      window.removeEventListener('afterprint', finish);
      root.classList.remove(PRINTING_CLASS);
    };
  }, []);

  const printedOn = new Intl.DateTimeFormat(`${i18n.language}-u-nu-latn`, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date());
  const commentDate = (iso: string): string => {
    const at = new Date(iso);
    return Number.isNaN(at.getTime())
      ? ''
      : new Intl.DateTimeFormat(`${i18n.language}-u-nu-latn`, {
          dateStyle: 'medium',
          timeStyle: 'short',
        }).format(at);
  };

  return createPortal(
    <article className="flow-task-print" dir={i18n.dir(i18n.language)} lang={i18n.language}>
      <header className="flow-task-print-header">
        <p className="flow-task-print-meta">
          {identifier} · {projectName}
        </p>
        <h1>{title}</h1>
      </header>

      <dl className="flow-task-print-fields">
        {fields.map((field) => (
          <div key={field.label}>
            <dt>{field.label}</dt>
            <dd>{field.value || '-'}</dd>
          </div>
        ))}
      </dl>

      {description.trim() && (
        <section>
          <h2>{t('flow.taskDetail.descriptionLabel')}</h2>
          <DescriptionEditor value={description} onSave={() => {}} placeholder="" disabled />
        </section>
      )}

      {subtasks.length > 0 && (
        <section>
          <h2>{t('flow.taskDetail.subtasks')}</h2>
          <ul className="flow-task-print-subtasks">
            {subtasks.map((subtask) => (
              <li key={subtask.id}>
                <span aria-hidden="true">{subtask.completed ? '☑' : '☐'}</span>{' '}
                <span className={subtask.completed ? 'flow-task-print-done' : undefined}>
                  {subtask.title}
                </span>
              </li>
            ))}
          </ul>
        </section>
      )}

      {comments.length > 0 && (
        <section>
          <h2>{t('flow.taskDetail.comments')}</h2>
          <ol className="flow-task-print-comments">
            {comments.map((comment) => (
              <li key={comment.id}>
                <p className="flow-task-print-meta">
                  <strong>{comment.author}</strong> · {commentDate(comment.createdAt)}
                </p>
                <p className="flow-task-print-body">{comment.body}</p>
              </li>
            ))}
          </ol>
        </section>
      )}

      <footer className="flow-task-print-meta">
        {t('flow.taskDetail.print.printedOn', { date: printedOn })}
      </footer>
    </article>,
    document.body
  );
};

export default TaskPrintView;
