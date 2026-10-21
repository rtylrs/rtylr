import type { HTMLAttributes, ReactNode } from 'react';
import { type RowProps, useVirtualRows } from '../hooks/useVirtualRows';

/**
 * Stands in for the rows a windowed table body isn't rendering (see
 * useVirtualRows). Renders nothing when there is no gap.
 */
export function VirtualSpacerRow({ height, columns }: { height: number; columns: number }) {
  if (height <= 0) return null;
  return (
    <tr aria-hidden="true" className="virtual-spacer" style={{ height }}>
      <td colSpan={columns} style={{ padding: 0, border: 0 }} />
    </tr>
  );
}

/** The same gap for a windowed list of plain elements. */
export function VirtualSpacer({ height }: { height: number }) {
  if (height <= 0) return null;
  return <div aria-hidden="true" className="virtual-spacer" style={{ height }} />;
}

interface VirtualListProps extends HTMLAttributes<HTMLDivElement> {
  count: number;
  /** Expected row height in px; rows are measured once rendered. */
  estimateSize: number;
  /** The CSS gap between rows, when the list is a flex or grid column. */
  gap?: number;
  /** Renders row `index`; spread `rowProps` on its outermost element. */
  renderRow: (index: number, rowProps: RowProps) => ReactNode;
}

/**
 * A div list that windows its rows once it gets long, for lists rendered
 * inline (a group in a grouped list, say) where a hook can't be called per
 * group. Every other prop lands on the wrapping div.
 */
export function VirtualList({ count, estimateSize, gap, renderRow, ...rest }: VirtualListProps) {
  const virtual = useVirtualRows(count, { estimateSize, gap });
  return (
    <div ref={virtual.listRef} {...rest}>
      <VirtualSpacer height={virtual.before} />
      {virtual.indexes.map((index) => renderRow(index, virtual.rowProps(index)))}
      <VirtualSpacer height={virtual.after} />
    </div>
  );
}
