import React, { useState } from 'react';
import { cdnImageUrl } from '@voxire/config';
import { BusinessReview } from '../../types';
import styles from './ReviewsPage.module.scss';

export interface ReviewSubmissionInput {
  authorName: string;
  rating: number;
  comment: string;
}

interface ReviewsPageProps {
  reviews: BusinessReview[];
  averageRating?: number;
  reviewCount?: number;
  previewMode?: boolean;
  onSubmitReview?: (input: ReviewSubmissionInput) => Promise<void> | void;
}

const StarRow: React.FC<{
  rating: number;
  max?: number;
  interactive?: boolean;
  onChange?: (r: number) => void;
}> = ({ rating, max = 5, interactive = false, onChange }) => {
  const [hovered, setHovered] = useState(0);
  return (
    <div className={styles.stars}>
      {Array.from({ length: max }).map((_, i) => {
        const val = i + 1;
        const filled = interactive ? val <= (hovered || rating) : val <= rating;
        return (
          <span
            key={i}
            className={filled ? styles.stars__filled : styles.stars__empty}
            onMouseEnter={interactive ? () => setHovered(val) : undefined}
            onMouseLeave={interactive ? () => setHovered(0) : undefined}
            onClick={interactive && onChange ? () => onChange(val) : undefined}
            style={interactive ? { cursor: 'pointer' } : undefined}
          >
            ★
          </span>
        );
      })}
    </div>
  );
};

const ReviewsPage: React.FC<ReviewsPageProps> = ({
  reviews,
  averageRating,
  reviewCount,
  previewMode = false,
  onSubmitReview,
}) => {
  const [newRating, setNewRating] = useState(0);
  const [newComment, setNewComment] = useState('');
  const [newName, setNewName] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitted, setSubmitted] = useState(false);

  const avg =
    averageRating ??
    (reviews.length > 0 ? reviews.reduce((s, r) => s + r.rating, 0) / reviews.length : 0);
  const total = reviewCount ?? reviews.length;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newRating || !newComment.trim()) return;
    setSubmitError(null);
    if (previewMode || !onSubmitReview) {
      setSubmitted(true);
      return;
    }
    setSubmitting(true);
    try {
      await onSubmitReview({
        authorName: newName.trim() || 'Guest',
        rating: newRating,
        comment: newComment.trim(),
      });
      setSubmitted(true);
      setNewRating(0);
      setNewName('');
      setNewComment('');
    } catch {
      setSubmitError('Could not submit your review right now. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className={styles.reviews}>
      {/* Rating summary */}
      <div className={styles.reviews__summary}>
        <span className={styles.reviews__avgNum}>{avg.toFixed(1)}</span>
        <div className={styles.reviews__avgRight}>
          <StarRow rating={Math.round(avg)} />
          <span className={styles.reviews__count}>
            {total} review{total !== 1 ? 's' : ''}
          </span>
        </div>
      </div>

      {/* Leave a review */}
      <div className={styles.reviews__form}>
        <h3 className={styles.reviews__formTitle}>Leave a Review</h3>
        {submitted ? (
          <p className={styles.reviews__thanks} role="status">
            {previewMode
              ? 'Preview only — guest reviews are saved for moderation on your live menu.'
              : 'Thanks for your review! It will appear once approved by the restaurant.'}
          </p>
        ) : (
          <form onSubmit={(e) => void handleSubmit(e)}>
            <div className={styles.reviews__formGroup}>
              <label className={styles.reviews__label}>Your Rating</label>
              <StarRow rating={newRating} interactive onChange={setNewRating} />
            </div>
            <div className={styles.reviews__formGroup}>
              <label className={styles.reviews__label}>Name</label>
              <input
                className={styles.reviews__input}
                type="text"
                placeholder="Your name"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
              />
            </div>
            <div className={styles.reviews__formGroup}>
              <label className={styles.reviews__label}>Comment</label>
              <textarea
                className={styles.reviews__textarea}
                placeholder="Tell us about your experience…"
                value={newComment}
                onChange={(e) => setNewComment(e.target.value)}
                rows={3}
              />
            </div>
            {submitError && (
              <p role="alert" style={{ color: '#dc2626', fontSize: '0.8rem', margin: '0 0 10px' }}>
                {submitError}
              </p>
            )}
            <button
              className={styles.reviews__submitBtn}
              type="submit"
              disabled={!newRating || !newComment.trim() || submitting}
            >
              {submitting ? 'Submitting…' : 'Submit Review'}
            </button>
          </form>
        )}
      </div>

      {/* Review list */}
      {reviews.length > 0 && (
        <div className={styles.reviews__list}>
          <h3 className={styles.reviews__listTitle}>Reviews</h3>
          {reviews.map((r) => {
            const avatarUrl = cdnImageUrl(r.avatar);
            return (
              <div key={r.id} className={styles.reviews__card}>
                <div className={styles.reviews__cardHeader}>
                  {avatarUrl ? (
                    <img src={avatarUrl} alt={r.author} className={styles.reviews__avatar} />
                  ) : (
                    <div className={styles.reviews__avatarFallback}>
                      {r.author.charAt(0).toUpperCase()}
                    </div>
                  )}
                  <div className={styles.reviews__cardMeta}>
                    <span className={styles.reviews__author}>{r.author}</span>
                    <div className={styles.reviews__cardStarRow}>
                      <StarRow rating={r.rating} />
                      <span className={styles.reviews__date}>{r.date}</span>
                    </div>
                  </div>
                </div>
                <p className={styles.reviews__comment}>{r.comment}</p>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default ReviewsPage;
