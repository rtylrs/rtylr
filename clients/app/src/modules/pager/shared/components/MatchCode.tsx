import React from 'react';

/**
 * Longer than a match number: a code from before the numbers ("PANDA2"),
 * which is shown smaller so it still fits.
 */
const MAX_SHORT_CODE = 4;

interface MatchCodeProps {
  code: string;
  className: string;
  /** Added when the code is longer than a match number. */
  longClassName?: string;
}

/**
 * The match code a phone shows the cashier, as the phone and the board both
 * render it. The server sends ASCII digits and they are shown exactly as sent,
 * left to right, also on the Arabic page, so the customer and the cashier read
 * the same characters whatever language each screen is in.
 */
const MatchCode = ({ code, className, longClassName }: MatchCodeProps): React.JSX.Element => (
  <bdi
    dir="ltr"
    translate="no"
    className={
      longClassName && code.length > MAX_SHORT_CODE ? `${className} ${longClassName}` : className
    }
  >
    {code}
  </bdi>
);

export default MatchCode;
