import { createContext, useContext } from 'react';
import type { Task } from '../../../shared/stores/flowStore';
import { DEFAULT_LIST_COLUMNS, type ListColumnId } from './listColumns';

/**
 * What every row of one List needs to agree on: which columns, how wide, and
 * which parents are open. Rows rendered on their own (a test, say) get the
 * List's defaults and keep their expansion to themselves.
 */
export interface ListGridState {
  columns: readonly ListColumnId[];
  widths: Partial<Record<ListColumnId, number>>;
  /** Null when no List owns the rows: each row then remembers its own. */
  expanded: ReadonlySet<string> | null;
  toggleExpanded: (taskId: string) => void;
  /** The List's current sort, applied to subtasks under their parent too. */
  sortRows: (tasks: Task[]) => Task[];
}

const STANDALONE: ListGridState = {
  columns: DEFAULT_LIST_COLUMNS,
  widths: {},
  expanded: null,
  toggleExpanded: () => {},
  sortRows: (tasks) => tasks,
};

export const ListGridContext = createContext<ListGridState>(STANDALONE);

export const useListGrid = (): ListGridState => useContext(ListGridContext);
