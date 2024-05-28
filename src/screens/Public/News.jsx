import React from "react";
import useAsync from "../../hooks/useAsync";

function News() {
  const [news, setNews] = React.useState([]);

  const getNews = useAsync({
    fn: configAp.UserCount,
    onSuccess: () => {
      setUsers(getUsersCount.result?.count);
    },
    onError: () => {
      alert(getUsersCount.error);
    },
    callOnMount: true,
  });
  return <div>News</div>;
}

export default News;
