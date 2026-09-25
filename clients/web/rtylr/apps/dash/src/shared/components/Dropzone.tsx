import React, { useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useDropzone } from 'react-dropzone';

interface DropzoneProps {
  onFileSelected: (file: File) => void;
  accept?: { [key: string]: string[] };
}

const Dropzone: React.FC<DropzoneProps> = ({ onFileSelected, accept = { 'image/*': [] } }) => {
  const { t } = useTranslation();
  const onDrop = useCallback(
    (acceptedFiles: File[]) => {
      if (acceptedFiles.length > 0) {
        onFileSelected(acceptedFiles[0]);
      }
    },
    [onFileSelected]
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept,
    multiple: false,
  });

  return (
    <div {...getRootProps({ className: 'dropzone' })} style={dropzoneStyle}>
      <input {...getInputProps()} />
      {isDragActive ? <p>{t('common.dropzone.active')}</p> : <p>{t('common.dropzone.idle')}</p>}
    </div>
  );
};

const dropzoneStyle: React.CSSProperties = {
  border: '2px dashed #9b51e0',
  borderRadius: '6px',
  padding: '20px',
  textAlign: 'center',
  cursor: 'pointer',
};

export default Dropzone;
