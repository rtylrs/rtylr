import React, { useEffect } from 'react';
import classNames from 'classnames';
import { ModalProps } from '../../types/common';
import './Modal.scss';

const maxWidthMap: Record<string, string> = {
  xs: '444px',
  sm: '600px',
  md: '900px',
  lg: '1200px',
  xl: '1536px',
};

const Modal: React.FC<ModalProps> = ({
  open,
  onClose,
  title,
  maxWidth = 'md',
  fullWidth = false,
  className,
  style,
  children,
}) => {
  useEffect(() => {
    if (open) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => { document.body.style.overflow = ''; };
  }, [open]);

  if (!open) return null;

  const handleBackdropClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) onClose();
  };

  return (
    <div className="voxire-modal-backdrop" onClick={handleBackdropClick}>
      <div
        className={classNames('voxire-modal', className)}
        style={{
          maxWidth: fullWidth ? '100%' : maxWidthMap[maxWidth],
          width: fullWidth ? '100%' : undefined,
          ...style,
        }}
        role="dialog"
        aria-modal="true"
      >
        {title && (
          <div className="voxire-modal__title">
            <span>{title}</span>
            <button
              type="button"
              className="voxire-modal__close"
              aria-label="Close"
              onClick={onClose}
            >
              ✕
            </button>
          </div>
        )}
        <div className="voxire-modal__content">{children}</div>
      </div>
    </div>
  );
};

export default Modal;
