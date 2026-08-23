import { useTranslation } from 'react-i18next';
import React from 'react';
import { GetStaticPaths, GetStaticProps } from 'next';
import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { blogPostingSchema } from '../../lib/structured-data';
import { getAllPosts, getPostBySlug, type Post, type Block } from '../../lib/blog/posts';
import styles from './blogPost.module.scss';

const renderBlock = (block: Block, i: number): React.ReactNode => {
  switch (block.type) {
    case 'h2':
      return (
        <h2 key={i} className={styles.h2}>
          {block.text}
        </h2>
      );
    case 'h3':
      return (
        <h3 key={i} className={styles.h3}>
          {block.text}
        </h3>
      );
    case 'p':
      return (
        <p key={i} className={styles.p}>
          {block.text}
        </p>
      );
    case 'quote':
      return (
        <blockquote key={i} className={styles.quote}>
          {block.text}
        </blockquote>
      );
    case 'ul':
      return (
        <ul key={i} className={styles.ul}>
          {block.items.map((item, j) => (
            <li key={j}>{item}</li>
          ))}
        </ul>
      );
    default:
      return null;
  }
};

interface BlogPostPageProps {
  post: Post;
}

const BlogPostPage = ({ post }: BlogPostPageProps): React.JSX.Element => {
  const { t } = useTranslation();
  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={post.metaTitle}
          description={post.description}
          canonical={`https://rtylr.com/blog/${post.slug}/`}
          ogType="article"
          ogDescription={post.description}
          structuredData={blogPostingSchema({
            title: post.title,
            slug: post.slug,
            description: post.description,
            datePublished: post.date,
          })}
        />

        <article className={styles.page}>
          <div className={styles.inner}>
            <Link href="/blog" className={styles.back}>
              <ArrowLeft size={15} /> {t('blog.post.allPosts')}
            </Link>
            <span className={styles.category}>{post.category}</span>
            <h1 className={styles.title}>{post.title}</h1>
            <div className={styles.meta}>
              <time dateTime={post.date}>{post.dateDisplay}</time>
              <span className={styles.dot}>·</span>
              <span>{t('blog.readingTime', { minutes: post.readingMinutes })}</span>
            </div>
            <div className={styles.body}>{post.blocks.map(renderBlock)}</div>

            <div className={styles.cta}>
              <h3>{t('blog.post.ctaTitle')}</h3>
              <p>{t('blog.post.ctaBody')}</p>
              <Link href="/demo" className={styles.ctaBtn}>
            {t('common.cta.bookDemo')}
          </Link>
            </div>
          </div>
        </article>
      </MainLayout>
    </AppLayout>
  );
};

export const getStaticPaths: GetStaticPaths = async () => ({
  paths: getAllPosts().map((p) => ({ params: { slug: p.slug } })),
  fallback: false,
});

export const getStaticProps: GetStaticProps<BlogPostPageProps> = async ({ params }) => {
  const post = getPostBySlug(params?.slug as string);
  if (!post) {
    return { notFound: true };
  }
  return { props: { post } };
};

export default BlogPostPage;
