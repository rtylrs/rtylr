import React, { useState } from 'react';
import { X, Package } from 'lucide-react';
import { cdnImageUrl } from '@voxire/config';

interface ImageGalleryProps {
  images: string[];
  name?: string;
}

const ImageGallery: React.FC<ImageGalleryProps> = ({ images, name = 'Product' }) => {
  const [selected, setSelected] = useState(0);
  const [lightbox, setLightbox] = useState(false);
  const selectedImageUrl = cdnImageUrl(images[selected]);

  if (!images || images.length === 0) {
    return (
      <div
        className="flex items-center justify-center h-64 rounded-xl"
        style={{ background: 'var(--hub-hover)', border: '1px solid var(--hub-border)' }}
      >
        <Package size={32} style={{ color: 'var(--hub-text-dim)' }} />
      </div>
    );
  }

  return (
    <>
      <div className="space-y-2">
        <button
          type="button"
          onClick={() => setLightbox(true)}
          className="w-full h-64 rounded-xl overflow-hidden"
          style={{ border: '1px solid var(--hub-border)' }}
        >
          {selectedImageUrl && (
            <img src={selectedImageUrl} alt={name} className="w-full h-full object-cover" />
          )}
        </button>
        {images.length > 1 && (
          <div className="flex gap-2">
            {images.map((img, i) => {
              const imageUrl = cdnImageUrl(img);
              return (
                <button
                  type="button"
                  key={i}
                  onClick={() => setSelected(i)}
                  className="w-16 h-16 rounded-lg overflow-hidden shrink-0"
                  style={{
                    border:
                      selected === i
                        ? '2px solid var(--erp-accent)'
                        : '1px solid var(--hub-border)',
                  }}
                >
                  {imageUrl && <img src={imageUrl} alt="" className="w-full h-full object-cover" />}
                </button>
              );
            })}
          </div>
        )}
      </div>
      {lightbox && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-6"
          style={{ background: 'var(--hub-overlay-opaque)' }}
          onClick={() => setLightbox(false)}
        >
          <button
            type="button"
            className="absolute top-4 right-4 h-8 w-8 rounded-full flex items-center justify-center"
            style={{ background: 'var(--hub-highlight)', color: 'var(--hub-text)' }}
          >
            <X size={16} />
          </button>
          {selectedImageUrl && (
            <img src={selectedImageUrl} alt={name} className="max-w-full max-h-full rounded-xl" />
          )}
        </div>
      )}
    </>
  );
};

export default ImageGallery;
