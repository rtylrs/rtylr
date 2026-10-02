import React from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { Trash2, Archive, Download, Edit3, X } from 'lucide-react';

interface BulkActionBarProps {
  count: number;
  onClear: () => void;
  onDelete?: () => void;
  onArchive?: () => void;
  onExport?: () => void;
  onEdit?: () => void;
}

const BulkActionBar: React.FC<BulkActionBarProps> = ({
  count,
  onClear,
  onDelete,
  onArchive,
  onExport,
  onEdit,
}) => {
  return (
    <AnimatePresence>
      {count > 0 && (
        <motion.div
          initial={{ y: 60, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          exit={{ y: 60, opacity: 0 }}
          transition={{ type: 'spring', stiffness: 400, damping: 30 }}
          className="fixed bottom-6 left-1/2 -translate-x-1/2 z-30 flex items-center gap-1 rounded-xl px-3 py-2"
          style={{
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
            boxShadow: 'var(--hub-shadow)',
          }}
        >
          <span
            className="text-xs font-semibold px-3 py-1 rounded-lg mr-1"
            style={{
              background: 'color-mix(in srgb, var(--erp-accent) 15%, transparent)',
              color: 'var(--erp-accent)',
            }}
          >
            {count} selected
          </span>
          {onEdit && (
            <button
              type="button"
              onClick={onEdit}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs"
              style={{ color: 'var(--hub-text-muted)' }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'transparent';
              }}
            >
              <Edit3 size={12} /> Edit
            </button>
          )}
          {onExport && (
            <button
              type="button"
              onClick={onExport}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs"
              style={{ color: 'var(--hub-text-muted)' }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'transparent';
              }}
            >
              <Download size={12} /> Export
            </button>
          )}
          {onArchive && (
            <button
              type="button"
              onClick={onArchive}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs"
              style={{ color: 'var(--hub-text-muted)' }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'transparent';
              }}
            >
              <Archive size={12} /> Archive
            </button>
          )}
          {onDelete && (
            <button
              type="button"
              onClick={onDelete}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs"
              style={{ color: 'var(--erp-danger)' }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.background =
                  'color-mix(in srgb, var(--erp-danger) 10%, transparent)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'transparent';
              }}
            >
              <Trash2 size={12} /> Delete
            </button>
          )}
          <button
            type="button"
            onClick={onClear}
            className="ml-1 h-7 w-7 rounded-lg flex items-center justify-center"
            style={{ color: 'var(--hub-text-dim)' }}
          >
            <X size={12} />
          </button>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default BulkActionBar;
