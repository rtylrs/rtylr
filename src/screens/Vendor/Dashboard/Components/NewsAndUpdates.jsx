import React from "react";
import adminApi from "../../../../api/admin";
import useAsync from "../../../../hooks/useAsync";
import { formatDate } from "../../../../utils/formatDate";

function NewsAndUpdates() {
  const [news, setNews] = React.useState([]);

  const getNews = useAsync({
    fn: adminApi.GetNews,
    onSuccess: () => {
      console.log(getNews.result);
      setNews(getNews?.result || []);
    },
    onError: () => {
      alert(getNews.error);
    },
    callOnMount: true,
  });

  return (
    <div className="news-container">
      {news &&
        news?.map((item) => (
          <div key={item.id} className="news-item">
            <img
              src={item?.image}
              alt={item.title}
              className="news-image"
              onError={(e) => {
                e.target.onerror = null;
                e.target.src = "https://via.placeholder.com/150";
              }}
            />
            <div className="news-content">
              <h3 className="news-title">{item.title}</h3>
              <p className="news-date">{formatDate(item.createdAt)}</p>
              <p className="news-description">
                {item.shortContent?.slice(0, 50)?.trim()}
                {item.shortContent && item.shortContent?.length > 50 && "....."}
              </p>
            </div>
          </div>
        ))}
    </div>
  );
}

export default NewsAndUpdates;
