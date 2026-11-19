import BlogPage from '../../components/blog/BlogPage';
import { blogStaticPaths, blogStaticProps } from '../../lib/blog/pages';

export const getStaticPaths = blogStaticPaths('en');
export const getStaticProps = blogStaticProps('en');

export default BlogPage;
