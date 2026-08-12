import React from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { X, Filter } from 'lucide-react';

interface FilterDrawerProps {
  open: boolean;
  onClose: () => void;
  onApply?: () => void;
  onReset?: () => void;
  title?: string;
  children: React.ReactNode;
}

const FilterDrawer: React.FC<FilterDrawerProps> = ({
  open,
  onClose,
  onApply,
  onReset,
  title = 'Filters',
  children,
}) => {
  return (
    <AnimatePresence>
      {open && (
        <>
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="fixed inset-0 z-40"
            style={{ background: 'var(--hub-overlay-light)' }}
          />
          <motion.aside
            initial={{ x: 320 }}
            animate={{ x: 0 }}
            exit={{ x: 320 }}
            transition={{ type: 'spring', stiffness: 300, damping: 30 }}
            className="fixed right-0 top-0 bottom-0 w-80 z-50 flex flex-col"
            style={{ background: 'var(--hub-card)', borderLeft: '1px solid var(--hub-border)' }}
          >
            <div
              className="flex items-center justify-between px-4 py-3"
              style={{ borderBottom: '1px solid var(--hub-border)' }}
            >
              <div className="flex items-center gap-2">
                <Filter size={14} style={{ color: 'var(--erp-accent)' }} />
                <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                  {title}
                </h3>
              </div>
              <button onClick={onClose}>
                <X size={14} style={{ color: 'var(--hub-text-muted)' }} />
              </button>
            </div>
            <div className="flex-1 overflow-y-auto p-4 space-y-4">{children}</div>
            <div className="p-4 flex gap-2" style={{ borderTop: '1px solid var(--hub-border)' }}>
              {onReset && (
                <button
                  onClick={onReset}
                  className="flex-1 px-3 py-2 rounded-lg text-xs font-medium"
                  style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
                >
                  Reset
                </button>
              )}
              <button
                onClick={() => {
                  onApply?.();
                  onClose();
                }}
                className="flex-1 px-3 py-2 rounded-lg text-xs font-semibold"
                style={{ background: 'var(--erp-accent)', color: 'var(--erp-accent-foreground)' }}
              >
                Apply
              </button>
            </div>
          </motion.aside>
        </>
      )}
    </AnimatePresence>
  );
};

export default FilterDrawer;
