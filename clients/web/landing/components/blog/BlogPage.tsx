import React from 'react';

import type { BlogPageProps } from '../../lib/blog/pages';
import AppLayout from '../Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import BlogListing from './BlogListing';
import BlogPost from './BlogPost';

/** Every blog route in both locales renders through this one component. */
const BlogPage = (props: BlogPageProps): React.JSX.Element => (
  <AppLayout>
    <MainLayout>
      {props.kind === 'post' ? (
        <BlogPost
          post={props.post}
          publishedLabel={props.publishedLabel}
          updatedLabel={props.updatedLabel}
          related={props.related}
          alternate={props.alternate}
          linkedTags={props.linkedTags}
        />
      ) : (
        <BlogListing {...props} />
      )}
    </MainLayout>
  </AppLayout>
);

export default BlogPage;
