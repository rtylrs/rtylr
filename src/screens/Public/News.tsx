import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import useAsync from "../../hooks/useAsync";
import { configApi } from "../../api";

const News: React.FC<{}> = () => {
  const { newsId } = useParams<{ newsId: string }>();
  const [news, setNews] = useState<any>({});

  const getNews = useAsync({
    fn: configApi.getSpecificNews,
    onSuccess: () => {
      setNews(getNews.result as any);
    },
    onError: () => {
      alert(getNews.error);
    },
    callOnMount: false,
  });

  useEffect(() => {
    if (newsId) {
      getNews.main({
        id: newsId,
      });
    }

     
  }, [newsId]);

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        textAlign: "center",
        padding: 40,
        minHeight: "60vh",
        maxWidth: 1350,
        margin: "0 auto",
      }}
    >
      {news ? (
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
        <p>Loading...</p>
      )}
    </div>
  );
};

export default News;
