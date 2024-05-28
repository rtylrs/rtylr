import React from "react";
import Skeleton from "react-loading-skeleton";
import "react-loading-skeleton/dist/skeleton.css";
import adminApi from "../../../../api/admin";
import useAsync from "../../../../hooks/useAsync";
import { formatDate } from "../../../../utils/formatDate";

function NewsAndUpdates() {
  const [news, setNews] = React.useState([]);
  const [loading, setLoading] = React.useState(true);

  const getNews = useAsync({
    fn: adminApi.GetNews,
    onSuccess: () => {
      setLoading(false);
      setNews(getNews?.result || []);
    },
    onError: () => {
      alert(getNews.error);
    },
    callOnMount: true,
  });

  return (
    <div className="news-container">
      {loading
        ? Array(2)
            .fill()
            .map((_, index) => (
              <div key={index} className="news-item">
                <Skeleton height={100} width={100} />
                <div
                  className="news-content"
                  style={{
                    marginLeft: "10px",
                  }}
                >
                  <h3 className="news-title">
                    <Skeleton width={200} />
                  </h3>
                  <p className="news-date">
                    <Skeleton width={100} />
                  </p>
                  <p className="news-description">
                    <Skeleton count={2} />
                  </p>
                </div>
              </div>
            ))
        : news?.map((item) => (
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
                  {item.shortContent &&
                    item.shortContent?.length > 50 &&
                    "....."}
                </p>
              </div>
            </div>
          ))}
    </div>
  );
}

export default NewsAndUpdates;
