import React from 'react';
import { Dialog, DialogTitle, DialogContent, IconButton } from '@mui/material';
import { Close as CloseIcon } from '@mui/icons-material';
import classNames from 'classnames';
import { ModalProps } from '../../types/common';
import './Modal.scss';

const Modal: React.FC<ModalProps> = ({
  open,
  onClose,
  title,
  maxWidth = 'md',
  fullWidth = false,
  className,
  style,
  children,
  ...props
}) => {
  const modalClasses = classNames(
    'voxire-modal',
    `voxire-modal--max-width-${maxWidth}`,
    className
  );

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth={maxWidth}
      fullWidth={fullWidth}
      className={modalClasses}
      style={style}
      {...props}
    >
      {title && (
        <DialogTitle className="voxire-modal__title">
          {title}
          <IconButton
            aria-label="close"
            onClick={onClose}
            className="voxire-modal__close-button"
          >
            <CloseIcon />
          </IconButton>
        </DialogTitle>
      )}

      <DialogContent className="voxire-modal__content">
        {children}
      </DialogContent>
    </Dialog>
  );
};

export default Modal;
