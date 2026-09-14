import React from 'react';
import { Trans } from 'react-i18next';

import { centsToDisplay, computeSeatProration } from '../hooks/queries/useSubscription';

interface SeatSavingsNoteProps {
  /** Monthly price of one seat, in minor units. Nothing renders at or below zero. */
  pricePerSeatCents: number;
}

/**
 * The proration line shown when deactivating a member: what this month's refund
 * comes to, and what the seat stops costing from next month.
 *
 * Extracted because Team and MemberDetail carried the same twelve lines of JSX,
 * and translating it meant a second copy of the same `Trans` block rather than
 * a second copy of a literal.
 *
 * `Trans` rather than two half-sentences: the two bold amounts sit mid-sentence
 * and Arabic orders them differently, so fragments would not translate.
 */
export function SeatSavingsNote({
  pricePerSeatCents,
}: SeatSavingsNoteProps): React.JSX.Element | null {
  if (pricePerSeatCents <= 0) return null;

  const { daysRemaining, daysInMonth, proratedCents, nextMonthName } =
    computeSeatProration(pricePerSeatCents);

  return (
    <span style={{ color: 'var(--dash-text-muted)', fontSize: '12px' }}>
      <Trans
        i18nKey="team.list.seatSavings"
        values={{
          saved: centsToDisplay(proratedCents),
          daysRemaining,
          daysInMonth,
          perSeat: centsToDisplay(pricePerSeatCents),
          month: nextMonthName,
        }}
        components={[
          <span key="saved" className="font-semibold" style={{ color: 'var(--dash-text)' }} />,
          <span key="perSeat" className="font-semibold" style={{ color: 'var(--dash-text)' }} />,
        ]}
      />
    </span>
  );
}

export default SeatSavingsNote;
