import BlogPage from '../../../components/blog/BlogPage';
import { blogStaticPaths, blogStaticProps } from '../../../lib/blog/pages';

/**
 * The Arabic blog. A real route rather than an entry in pages/[locale],
 * because Arabic ships for this section only: see `sections` in
 * lib/locales.ts. No post files in Arabic means no paths, so nothing under
 * /ar/ is exported until the first Arabic post exists.
 */
export const getStaticPaths = blogStaticPaths('ar');
export const getStaticProps = blogStaticProps('ar');

export default BlogPage;
