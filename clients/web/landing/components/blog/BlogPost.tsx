import Link from 'next/link';
import React from 'react';
import { useTranslation } from 'react-i18next';

import { cn } from '../../lib/cn';
import { blogPostingSchema } from '../../lib/structured-data';
import {
  authorById,
  blogPath,
  segmentById,
  type Block,
  type ListedPost,
  type Post,
} from '../../lib/blog/types';
import { tagLabel } from '../../lib/blog/tags';
import HelmetContainer, { ogImageFor } from '../HelmetContainer';
import BlogCta from './BlogCta';
import { Inline, plainText } from './Inline';
import PostCard from './PostCard';
import styles from './Blog.module.scss';

/** A table of contents earns its space from this many sections up. */
const TOC_MIN_SECTIONS = 4;

const sectionId = (n: number): string => `section-${n}`;

/** "Abed El-Fattah Amouneh" -> "AA". First and last word, so it works in Arabic too. */
const initials = (name: string): string => {
  const words = name.split(/\s+/).filter(Boolean);
  return words.length ? `${words[0][0]}${words.length > 1 ? words[words.length - 1][0] : ''}` : '';
};

/**
 * Where the first call to action goes when a post does not place one: right
 * before its second section, which is after the first useful answer and
 * well before the end.
 */
function withInlineCta(blocks: Block[]): Block[] {
  if (blocks.some((b) => b.type === 'cta')) return blocks;
  const h2s = blocks.flatMap((b, i) => (b.type === 'h2' ? [i] : []));
  if (h2s.length < 2) return blocks;
  return [...blocks.slice(0, h2s[1]), { type: 'cta' }, ...blocks.slice(h2s[1])];
}

interface Props {
  post: Post;
  publishedLabel: string;
  updatedLabel: string | null;
  related: ListedPost[];
  alternate: string | null;
  linkedTags: string[];
}

export function BlogPost({
  post,
  publishedLabel,
  updatedLabel,
  related,
  alternate,
  linkedTags,
}: Props): React.JSX.Element {
  const { t } = useTranslation();
  const author = authorById(post.author);
  const authorKey = author?.key ?? 'abed';
  const authorPath = blogPath(post.locale, `author/${post.author}/`);
  const segment = segmentById(post.segment);
  const segmentLabel = t(`blog.segments.${post.segment}.label`);
  const englishPath = blogPath('en', `${post.slug}/`);
  const canonical = `https://rtylr.com${englishPath}`;
  const localisedPath = blogPath(post.locale, `${post.slug}/`);
  const image = ogImageFor(`https://rtylr.com${localisedPath}`);

  const blocks = withInlineCta(post.blocks);
  const sections = blocks.filter((b): b is { type: 'h2'; text: string } => b.type === 'h2');
  let h2Seen = 0;

  const renderBlock = (block: Block, i: number): React.ReactNode => {
    switch (block.type) {
      case 'h2':
        h2Seen += 1;
        return (
          <h2 key={i} id={sectionId(h2Seen)} className={styles.h2}>
            <Inline text={block.text} />
          </h2>
        );
      case 'h3':
        return (
          <h3 key={i} className={styles.h3}>
            <Inline text={block.text} />
          </h3>
        );
      case 'p':
        return (
          <p key={i}>
            <Inline text={block.text} />
          </p>
        );
      case 'quote':
        return (
          <blockquote key={i} className={styles.quote}>
            <Inline text={block.text} />
          </blockquote>
        );
      case 'ul':
      case 'ol': {
        const List = block.type;
        return (
          <List key={i} className={styles.list}>
            {block.items.map((item, j) => (
              <li key={j}>
                <Inline text={item} />
              </li>
            ))}
          </List>
        );
      }
      case 'table':
        return (
          <div key={i} className={styles.tableWrap}>
            <table className={styles.table}>
              {block.caption && (
                <caption>
                  <Inline text={block.caption} />
                </caption>
              )}
              <thead>
                <tr>
                  {block.head.map((cell, j) => (
                    <th key={j} scope="col">
                      <Inline text={cell} />
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {block.rows.map((row, r) => (
                  <tr key={r}>
                    {row.map((cell, j) =>
                      j === 0 ? (
                        <th key={j} scope="row">
                          <Inline text={cell} />
                        </th>
                      ) : (
                        <td key={j}>
                          <Inline text={cell} />
                        </td>
                      )
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        );
      case 'cta':
        return <BlogCta key={i} segment={post.segment} placement="inline" slug={post.slug} />;
      default:
        return null;
    }
  };

  return (
    <>
      <HelmetContainer
        title={post.metaTitle}
        description={post.description}
        canonical={canonical}
        ogType="article"
        image={image}
        structuredData={blogPostingSchema({
          title: post.title,
          path: localisedPath,
          description: post.description,
          datePublished: post.date,
          dateModified: post.dateModified,
          inLanguage: post.locale,
          section: segmentLabel,
          keywords: [post.keyword, ...post.tags.map((tag) => tagLabel(tag, post.locale))],
          image,
          author: {
            name: t(`blog.authors.${authorKey}.name`),
            path: authorPath,
            sameAs: author?.links.map((l) => l.href) ?? [],
          },
          breadcrumbs: [
            { label: 'rtylr', path: '/' },
            { label: t('blog.post.allPosts'), path: blogPath(post.locale) },
            { label: segmentLabel, path: blogPath(post.locale, `${segment.slug}/`) },
            { label: post.title, path: localisedPath },
          ],
        })}
      />

      <article className={cn(styles.post, styles[`seg_${post.segment}`])}>
        <header className={styles.postHeader}>
          <nav aria-label={t('blog.post.breadcrumb')} className={styles.crumbs}>
            <Link href={blogPath(post.locale)}>{t('blog.post.allPosts')}</Link>
            <span aria-hidden="true">/</span>
            <Link href={blogPath(post.locale, `${segment.slug}/`)}>{segmentLabel}</Link>
          </nav>
          <span className={styles.label}>{segmentLabel}</span>
          <h1 className={styles.postTitle}>{post.title}</h1>
          <p className={styles.postExcerpt}>{post.excerpt}</p>
          <p className={styles.postMeta}>
            <span className={styles.avatar} aria-hidden="true">
              {initials(t(`blog.authors.${authorKey}.name`))}
            </span>
            <span>
              {t('blog.post.by')} <Link href={authorPath}>{t(`blog.authors.${authorKey}.name`)}</Link>
            </span>
            <span aria-hidden="true">·</span>
            <time dateTime={post.date}>{t('blog.post.published', { date: publishedLabel })}</time>
            {updatedLabel && post.dateModified && (
              <>
                <span aria-hidden="true">·</span>
                <time dateTime={post.dateModified}>
                  {t('blog.post.updated', { date: updatedLabel })}
                </time>
              </>
            )}
            <span aria-hidden="true">·</span>
            <span>{t('blog.readingTime', { minutes: post.readingMinutes })}</span>
          </p>
          {alternate && (
            <Link
              href={alternate}
              className={styles.otherLanguage}
              hrefLang={post.locale === 'en' ? 'ar' : 'en'}
              lang={post.locale === 'en' ? 'ar' : 'en'}
            >
              {t('blog.post.otherLanguage')}
            </Link>
          )}
        </header>

        <div className={styles.postLayout}>
          {sections.length >= TOC_MIN_SECTIONS && (
            <nav className={styles.toc} aria-labelledby="toc-title">
              <p id="toc-title" className={styles.tocTitle}>
                {t('blog.post.toc')}
              </p>
              <ol>
                {sections.map((s, i) => (
                  <li key={i}>
                    <a href={`#${sectionId(i + 1)}`}>{plainText(s.text)}</a>
                  </li>
                ))}
              </ol>
            </nav>
          )}

          <div className={styles.postBody}>
            {blocks.map(renderBlock)}

            {post.faq && post.faq.length > 0 && (
              <section className={styles.faq}>
                <h2 className={styles.h2}>{t('blog.post.faq')}</h2>
                {post.faq.map((item, i) => (
                  <div key={i} className={styles.faqItem}>
                    <h3 className={styles.h3}>{item.question}</h3>
                    <p>
                      <Inline text={item.answer} />
                    </p>
                  </div>
                ))}
              </section>
            )}

            {post.tags.length > 0 && (
              <p className={styles.tags}>
                <span>{t('blog.tag.label')}:</span>
                {post.tags.map((tag) =>
                  linkedTags.includes(tag) ? (
                    <Link key={tag} href={blogPath(post.locale, `tag/${tag}/`)} className={styles.tag}>
                      {tagLabel(tag, post.locale)}
                    </Link>
                  ) : (
                    <span key={tag} className={styles.tag}>
                      {tagLabel(tag, post.locale)}
                    </span>
                  )
                )}
              </p>
            )}

            <div className={styles.authorBox}>
              <p className={styles.authorEyebrow}>{t('blog.post.aboutAuthor')}</p>
              <p className={styles.authorName}>
                <Link href={authorPath}>{t(`blog.authors.${authorKey}.name`)}</Link>
                <span>{t(`blog.authors.${authorKey}.role`)}</span>
              </p>
              <p className={styles.authorBio}>{t(`blog.authors.${authorKey}.bio`)}</p>
            </div>

            <BlogCta segment={post.segment} placement="end" slug={post.slug} />
          </div>
        </div>
      </article>

      {related.length > 0 && (
        <section className={styles.related}>
          <h2 className={styles.sectionTitle}>{t('blog.post.related')}</h2>
          <div className={styles.grid}>
            {related.map((p) => (
              <PostCard key={p.slug} post={p} />
            ))}
          </div>
        </section>
      )}
    </>
  );
}

export default BlogPost;
