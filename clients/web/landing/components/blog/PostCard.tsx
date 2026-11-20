import Link from 'next/link';
import React from 'react';
import { useTranslation } from 'react-i18next';

import { cn } from '../../lib/cn';
import { postPath, type ListedPost } from '../../lib/blog/types';
import Cover from './Cover';
import styles from './Blog.module.scss';

/**
 * lead: the index's top story, a full colour block. An h2, because it sits
 *       before any section heading.
 * wide: spans two grid columns.
 */
export type CardVariant = 'lead' | 'wide' | 'normal';

export function PostCard({
  post,
  variant = 'normal',
}: {
  post: ListedPost;
  variant?: CardVariant;
}): React.JSX.Element {
  const { t } = useTranslation();
  const Heading = variant === 'lead' ? 'h2' : 'h3';
  return (
    <Link
      href={postPath(post)}
      className={cn(
        styles.card,
        variant === 'lead' && styles.leadCard,
        variant === 'wide' && styles.wideCard,
        styles[`seg_${post.segment}`]
      )}
    >
      <Cover cover={post.cover} />
      <div className={styles.cardBody}>
        <span className={styles.label}>{t(`blog.segments.${post.segment}.label`)}</span>
        <Heading className={styles.cardTitle}>{post.title}</Heading>
        <p className={styles.cardExcerpt}>{post.excerpt}</p>
        <p className={styles.cardMeta}>
          <time dateTime={post.dateModified ?? post.date}>{post.dateLabel}</time>
          <span aria-hidden="true"> · </span>
          {t('blog.readingTimeShort', { minutes: post.readingMinutes })}
        </p>
      </div>
    </Link>
  );
}

export default PostCard;
