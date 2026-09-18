export const CARET_FOCUS_COMPOSER_EVENT = 'caret:focus-composer';

export function requestCaretComposerFocus(): void {
  const dispatch = (): void => {
    window.dispatchEvent(new Event(CARET_FOCUS_COMPOSER_EVENT));
  };

  window.setTimeout(dispatch, 0);
  window.setTimeout(dispatch, 80);
}
