import React from 'react';
import Link from 'next/link';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { Clock } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { blurFadeIn, staggerBlur } from '../../lib/animations';
import { blogIndexSchema } from '../../lib/structured-data';
import { getAllPosts } from '../../lib/blog/posts';
import styles from './blog.module.scss';

const BlogPage = (): React.JSX.Element => {
  const { t } = useTranslation();
  const posts = getAllPosts();
  const [featured, ...rest] = posts;

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('blog.meta.title')}
          description={t('blog.meta.description')}
          canonical="https://rtylr.com/blog/"
          structuredData={blogIndexSchema({
            description:
              t('blog.meta.schemaDescription'),
            posts: posts.map((p) => ({
              title: p.title,
              slug: p.slug,
              datePublished: p.date,
            })),
          })}
        />

        <div className={styles.page}>
          <section className={styles.hero}>
            <motion.div
              className={styles.heroInner}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
            {t('blog.hero.eyebrow')}
          </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
            {t('blog.hero.title')}
          </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                {t('blog.hero.subtitle')}
              </motion.p>
              <motion.p className={styles.heroIntro} variants={blurFadeIn}>
                {t('blog.hero.intro')}
              </motion.p>
            </motion.div>
          </section>

          {featured && (
            <section className={styles.featuredSection}>
              <Link href={`/blog/${featured.slug}`} className={styles.featured}>
                <div className={styles.featuredLeft}>
                  <span className={styles.featuredBadge}>{featured.category.toUpperCase()}</span>
                  <h2 className={styles.featuredTitle}>{featured.title}</h2>
                  <p className={styles.featuredExcerpt}>{featured.excerpt}</p>
                  <div className={styles.featuredMeta}>
                    <time dateTime={featured.date}>{featured.dateDisplay}</time>
                    <span className={styles.postDot}>·</span>
                    <span className={styles.postReadTime}>
                      <Clock size={13} /> {t('blog.readingTime', { minutes: featured.readingMinutes })}
                    </span>
                  </div>
                </div>
                <div className={styles.featuredRight}>
                  <div className={styles.featuredArtwork}>
                    <div className={styles.artworkBar} style={{ width: '90%' }} />
                    <div className={styles.artworkBar} style={{ width: '70%' }} />
                    <div className={styles.artworkBar} style={{ width: '80%' }} />
                  </div>
                </div>
              </Link>
            </section>
          )}

          {rest.length > 0 && (
            <section className={styles.gridSection}>
              <div className={styles.grid}>
                {rest.map((post) => (
                  <Link key={post.slug} href={`/blog/${post.slug}`} className={styles.postCard}>
                    <div className={styles.postMetaTop}>
                      <span className={styles.postCategory}>{post.category}</span>
                    </div>
                    <h3 className={styles.postTitle}>{post.title}</h3>
                    <p className={styles.postExcerpt}>{post.excerpt}</p>
                    <div className={styles.postMetaBottom}>
                      <time dateTime={post.date}>{post.dateDisplay}</time>
                      <span className={styles.postDot}>·</span>
                      <span className={styles.postReadTime}>
                        <Clock size={12} /> {t('blog.readingTimeShort', { minutes: post.readingMinutes })}
                      </span>
                    </div>
                  </Link>
                ))}
              </div>
            </section>
          )}
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default BlogPage;
