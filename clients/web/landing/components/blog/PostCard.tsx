import Link from 'next/link';
import React from 'react';
import { useTranslation } from 'react-i18next';

import { postPath, type ListedPost } from '../../lib/blog/types';
import Cover from './Cover';
import styles from './Blog.module.scss';

export function PostCard({
  post,
  featured = false,
}: {
  post: ListedPost;
  featured?: boolean;
}): React.JSX.Element {
  const { t } = useTranslation();
  const Heading = featured ? 'h2' : 'h3';
  return (
    <Link href={postPath(post)} className={featured ? styles.featured : styles.card}>
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
