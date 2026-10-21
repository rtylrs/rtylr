import React from 'react';

export interface BarItem {
  id: string;
  label: string;
  /** What the bar's length encodes. */
  value: number;
  /** The value as shown at the bar's tip. */
  display: string;
  hint?: string;
  onClick?: () => void;
}

/**
 * Horizontal bars in one color, the longest at full length, each with its
 * value at the tip and any detail under its label. Rows open what they count.
 */
export const BarList: React.FC<{ items: BarItem[]; label: string }> = ({ items, label }) => {
  const max = Math.max(0, ...items.map((item) => item.value));
  // Room kept at the end of every track for the longest value at its tip.
  const reserve = Math.max(0, ...items.map((item) => item.display.length)) + 2;
  return (
    <ul className="crm-dash-bars" aria-label={label}>
      {items.map((item) => (
        <li key={item.id}>
          <button type="button" className="crm-dash-bar-row" onClick={item.onClick}>
            <span className="crm-dash-bar-label">
              {item.label}
              {item.hint && <span className="crm-dash-bar-hint">{item.hint}</span>}
            </span>
            <span className="crm-dash-bar-track">
              {item.value > 0 && max > 0 && (
                <span
                  className="crm-dash-bar"
                  style={{ width: `calc((100% - ${reserve}ch) * ${item.value / max})` }}
                />
              )}
              <span className="crm-dash-bar-value">{item.display}</span>
            </span>
          </button>
        </li>
      ))}
    </ul>
  );
};
