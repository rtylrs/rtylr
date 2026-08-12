import React from 'react';
import { motion } from 'motion/react';
import { X } from 'lucide-react';

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
  width?: number;
  footer?: React.ReactNode;
}

const Modal: React.FC<ModalProps> = ({ open, onClose, title, children, width = 480, footer }) => {
  if (!open) return null;
  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 50,
        background: 'var(--finance-overlay-strong)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backdropFilter: 'blur(4px)',
      }}
      onClick={onClose}
    >
      <motion.div
        initial={{ opacity: 0, scale: 0.95, y: 12 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        transition={{ duration: 0.2 }}
        onClick={(e) => e.stopPropagation()}
        style={{
          background: 'var(--hub-surface)',
          border: '1px solid var(--hub-border)',
          borderRadius: 16,
          padding: 24,
          width,
          maxWidth: '90vw',
          maxHeight: '90vh',
          overflow: 'auto',
          boxShadow: '0 40px 80px var(--finance-overlay-medium)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: 20,
          }}
        >
          <h2 style={{ margin: 0, fontSize: 16, fontWeight: 700, color: 'var(--hub-text)' }}>
            {title}
          </h2>
          <button
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--hub-text-muted)',
              padding: 4,
            }}
          >
            <X size={18} />
          </button>
        </div>
        <div>{children}</div>
        {footer && (
          <div style={{ display: 'flex', gap: 10, marginTop: 24, justifyContent: 'flex-end' }}>
            {footer}
          </div>
        )}
      </motion.div>
    </div>
  );
};

export default Modal;
