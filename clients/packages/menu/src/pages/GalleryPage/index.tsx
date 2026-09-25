import React, { useState } from 'react';
import { cdnImageUrl } from '@voxire/config';
import styles from './GalleryPage.module.scss';

interface GalleryPageProps {
  photos: string[];
  businessName: string;
}

const GalleryPage: React.FC<GalleryPageProps> = ({ photos, businessName }) => {
  const [lightbox, setLightbox] = useState<string | null>(null);

  if (!photos || photos.length === 0) {
    return (
      <div className={styles.gallery__empty}>
        <span className={styles.gallery__emptyIcon}>📷</span>
        <p>No photos yet.</p>
      </div>
    );
  }

  return (
    <div className={styles.gallery}>
      <div className={styles.gallery__grid}>
        {photos.map((src, i) => {
          const imageUrl = cdnImageUrl(src);
          return (
            <button
              key={i}
              className={styles.gallery__item}
              onClick={() => setLightbox(src)}
              aria-label={`View photo ${i + 1}`}
            >
              {imageUrl && (
                <img src={imageUrl} alt={`${businessName} photo ${i + 1}`} loading="lazy" />
              )}
            </button>
          );
        })}
      </div>

      {/* Lightbox */}
      {lightbox && (
        <div className={styles.gallery__lightbox} onClick={() => setLightbox(null)}>
          <button className={styles.gallery__lightboxClose} onClick={() => setLightbox(null)}>
            ✕
          </button>
          <img
            src={cdnImageUrl(lightbox)}
            alt="Enlarged view"
            className={styles.gallery__lightboxImg}
            onClick={(e) => e.stopPropagation()}
          />
        </div>
      )}
    </div>
  );
};

export default GalleryPage;
