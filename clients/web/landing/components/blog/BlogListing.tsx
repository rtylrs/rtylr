import Link from 'next/link';
import React from 'react';
import { useTranslation } from 'react-i18next';

import { blogCollectionSchema } from '../../lib/structured-data';
import type { BlogPageProps } from '../../lib/blog/pages';
import { tagLabel } from '../../lib/blog/tags';
import {
  authorById,
  blogPath,
  postPath,
  segmentById,
  type BlogLocale,
  type SegmentId,
} from '../../lib/blog/types';
import HelmetContainer from '../HelmetContainer';
import PostCard from './PostCard';
import styles from './Blog.module.scss';

type ListingProps = Exclude<BlogPageProps, { kind: 'post' }>;

function SegmentNav({
  locale,
  segments,
  current,
}: {
  locale: BlogLocale;
  segments: SegmentId[];
  current?: SegmentId;
}): React.JSX.Element {
  const { t } = useTranslation();
  return (
    <nav className={styles.segmentNav} aria-label={t('blog.sections.navLabel')}>
      <Link href={blogPath(locale)} aria-current={current ? undefined : 'page'}>
        {t('blog.sections.all')}
      </Link>
      {segments.map((id) => (
        <Link
          key={id}
          href={blogPath(locale, `${segmentById(id).slug}/`)}
          aria-current={current === id ? 'page' : undefined}
        >
          {t(`blog.segments.${id}.label`)}
        </Link>
      ))}
    </nav>
  );
}

export function BlogListing(props: ListingProps): React.JSX.Element {
  const { t } = useTranslation();
  const { locale, posts, segments } = props;

  // Everything that differs between the four listing kinds, in one place.
  let rest = '';
  let eyebrow = '';
  let title = '';
  let intro: React.ReactNode = null;
  let metaTitle = '';
  let metaDescription = '';
  let heading = t('blog.sections.latest');
  let crumb = '';
  switch (props.kind) {
    case 'index':
      eyebrow = t('blog.hero.eyebrow');
      title = t('blog.hero.title');
      metaTitle = t('blog.meta.title');
      metaDescription = t('blog.meta.description');
      intro = (
        <>
          <p className={styles.heroSubtitle}>{t('blog.hero.subtitle')}</p>
          <p className={styles.heroIntro}>{t('blog.hero.intro')}</p>
        </>
      );
      break;
    case 'segment':
      rest = `${segmentById(props.segment).slug}/`;
      eyebrow = t(`blog.segments.${props.segment}.label`);
      title = t(`blog.segments.${props.segment}.title`);
      metaTitle = t(`blog.segments.${props.segment}.metaTitle`);
      metaDescription = t(`blog.segments.${props.segment}.metaDescription`);
      intro = <p className={styles.heroSubtitle}>{t(`blog.segments.${props.segment}.intro`)}</p>;
      crumb = eyebrow;
      break;
    case 'tag': {
      const label = tagLabel(props.tag, locale);
      rest = `tag/${props.tag}/`;
      eyebrow = t('blog.tag.eyebrow');
      title = label;
      metaTitle = t('blog.tag.metaTitle', { tag: label });
      metaDescription = t('blog.tag.metaDescription', { tag: label });
      intro = <p className={styles.heroSubtitle}>{t('blog.tag.intro', { tag: label })}</p>;
      heading = t('blog.tag.heading', { tag: label });
      crumb = label;
      break;
    }
    case 'author': {
      const author = authorById(props.author);
      const key = author?.key ?? 'abed';
      const name = t(`blog.authors.${key}.name`);
      rest = `author/${props.author}/`;
      eyebrow = t('blog.authors.eyebrow');
      title = name;
      metaTitle = t(`blog.authors.${key}.metaTitle`);
      metaDescription = t(`blog.authors.${key}.metaDescription`);
      intro = (
        <>
          <p className={styles.heroRole}>{t(`blog.authors.${key}.role`)}</p>
          <p className={styles.heroSubtitle}>{t(`blog.authors.${key}.bio`)}</p>
          {author && author.links.length > 0 && (
            <p className={styles.authorLinks}>
              <span>{t('blog.authors.links')}:</span>
              {author.links.map((l) => (
                <a key={l.href} href={l.href} rel="me noopener" target="_blank">
                  {l.label}
                </a>
              ))}
            </p>
          )}
        </>
      );
      heading = t('blog.authors.heading', { name });
      crumb = name;
      break;
    }
  }

  const path = blogPath(locale, rest);
  const [featured, ...others] = props.kind === 'index' ? posts : [];
  const grid = props.kind === 'index' ? others : posts;

  return (
    <>
      <HelmetContainer
        title={metaTitle}
        description={metaDescription}
        canonical={`https://rtylr.com${blogPath('en', rest)}`}
        structuredData={blogCollectionSchema({
          name: metaTitle,
          path,
          description: metaDescription,
          inLanguage: locale,
          posts: posts.map((p) => ({ title: p.title, path: postPath(p), datePublished: p.date })),
          breadcrumbs: [
            { label: 'rtylr', path: '/' },
            { label: t('blog.post.allPosts'), path: blogPath(locale) },
            ...(crumb ? [{ label: crumb, path }] : []),
          ],
        })}
      />

      <div className={styles.page}>
        <section className={styles.hero}>
          <span className={props.kind === 'segment' ? styles.label : styles.eyebrow}>{eyebrow}</span>
          <h1 className={styles.heroTitle}>{title}</h1>
          {intro}
          {props.alternate && (
            <Link
              href={props.alternate}
              className={styles.otherLanguage}
              hrefLang={locale === 'en' ? 'ar' : 'en'}
              lang={locale === 'en' ? 'ar' : 'en'}
            >
              {t('blog.sections.otherLanguage')}
            </Link>
          )}
          <SegmentNav
            locale={locale}
            segments={segments}
            current={props.kind === 'segment' ? props.segment : undefined}
          />
        </section>

        {featured && (
          <section className={styles.featuredSection}>
            <PostCard post={featured} featured />
          </section>
        )}

        {props.kind === 'index' && segments.length > 0 && (
          <section className={styles.segments}>
            <h2 className={styles.sectionTitle}>{t('blog.sections.title')}</h2>
            <div className={styles.segmentGrid}>
              {segments.map((id) => (
                <Link
                  key={id}
                  href={blogPath(locale, `${segmentById(id).slug}/`)}
                  className={styles.segmentCard}
                >
                  <span className={styles.label}>{t(`blog.segments.${id}.label`)}</span>
                  <h3 className={styles.cardTitle}>{t(`blog.segments.${id}.title`)}</h3>
                  <p className={styles.cardExcerpt}>{t(`blog.segments.${id}.intro`)}</p>
                  <span className={styles.more}>{t('blog.sections.browse')}</span>
                </Link>
              ))}
            </div>
          </section>
        )}

        {grid.length > 0 && (
          <section className={styles.gridSection}>
            <h2 className={styles.sectionTitle}>{heading}</h2>
            <div className={styles.grid}>
              {grid.map((post) => (
                <PostCard key={post.slug} post={post} />
              ))}
            </div>
          </section>
        )}
      </div>
    </>
  );
}

export default BlogListing;
