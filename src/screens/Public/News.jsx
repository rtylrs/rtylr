import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import useAsync from "../../hooks/useAsync";
import configApi from "../../api/config";

function News() {
  const { newsId } = useParams();
  const [news, setNews] = useState(null);

  const getNews = useAsync({
    fn: configApi.getSpecificNews,
    onSuccess: () => {
      setNews(getNews.result);
    },
    onError: () => {
      alert(getNews.error);
    },
    callOnMount: false,
  });

  useEffect(() => {
    if (newsId) {
      console.log(newsId);
      getNews.main({
        id: newsId,
      });
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [newsId]);

  return (
    <div>
      <h1>News</h1>
      {news ? (
        <div>
          <h2>{news.title}</h2>
          <div dangerouslySetInnerHTML={{ __html: news.content }} />
          <img
            src={news.image}
            alt={news.title}
            style={{ height: 200, width: 200 }}
          />
        </div>
      ) : (
        <p>Loading...</p>
      )}
    </div>
  );
}

export default News;
