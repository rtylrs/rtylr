import React from 'react';
import { useRouter } from 'next/router';
import { useQuery } from '@tanstack/react-query';
import { configApi } from '../api';

const News: React.FC<{}> = () => {
  const router = useRouter();
  const { newsId } = router.query;

  const { data: news, isLoading, error } = useQuery({
    queryKey: ['news', newsId],
    queryFn: () => configApi.getSpecificNews({ id: newsId as string }),
    enabled: !!newsId,
  });

  if (error) {
    return <div>Error loading news</div>;
  }

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        textAlign: 'center',
        padding: 40,
        minHeight: '60vh',
        maxWidth: 1350,
        margin: '0 auto',
      }}
    >
      {isLoading ? (
        <p>Loading...</p>
      ) : news ? (
        <div>
          <h2>{news.title}</h2>
          <div dangerouslySetInnerHTML={{ __html: news.content }} />
          {/* Uncomment if you need to display the image
          <img
            src={news.image}
            alt={news.title}
            style={{ height: 200, width: 200 }}
          /> */}
        </div>
      ) : (
        <p>News not found</p>
      )}
    </div>
  );
};

export default News;
