import React from 'react';

interface LinesProps {
  /** Text whose `\n` characters are deliberate line breaks. */
  text: string;
}

/**
 * Render a string's newlines as real line breaks.
 *
 * Headings on this site are deliberately broken across lines, and that used
 * to be done with a literal `<br />` in the markup. Moving the copy to the
 * locale files turned those into `\n` inside a string, and JSX collapses a
 * newline in text content to a single space: the build output went from
 * two lines to `What does a register\nhave to get right?` rendered as one.
 *
 * Nothing errored and nothing looked wrong in a diff. Only the rendered
 * page changed, which is why `tests/unit/i18n-keys.test.mjs` now asserts
 * that any locale value containing `\n` is passed either to this component
 * or to SplitWords, which does its own splitting.
 *
 * Use it for headings and other short display copy. Do not use it to fake
 * paragraphs: two paragraphs are two keys.
 */
const Lines = ({ text }: LinesProps): React.JSX.Element => (
  <>
    {text.split('\n').map((line, i) => (
      <React.Fragment key={`${i}-${line}`}>
        {i > 0 && <br />}
        {line}
      </React.Fragment>
    ))}
  </>
);

export default Lines;
