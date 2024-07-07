import Skeleton from "react-loading-skeleton";
import "react-loading-skeleton/dist/skeleton.css";
import { formatDate } from "../../../../utils/formatDate";
import { Link } from "react-router-dom";
import { useSelector } from "react-redux";
import { RootState } from "../../../../redux/store";

function NewsAndUpdates() {
  const news = useSelector((state: RootState) => state.config.news);
  const loading = useSelector((state: RootState) => state.config.loading);

  return (
    <div className="news-container">
      {loading
        ? Array(2)
            .fill({
              title: "",
              date: "",
              description: "",
            })
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
            <Link
              key={item._id}
              className="news-item"
              to={item.isDetailed && `/news/${item._id}`}
            >
              <img
                src={item?.image}
                alt={item.title}
                className="news-image"
                onError={(e) => {
                  const target = e.target as HTMLImageElement;
                  target.onerror = null;
                  target.src = "https://via.placeholder.com/150";
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
            </Link>
          ))}
    </div>
  );
}

export default NewsAndUpdates;
