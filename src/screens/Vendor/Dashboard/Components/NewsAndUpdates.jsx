import React from "react";

function NewsAndUpdates() {
  const newsItems = [
    {
      id: 1,
      image: "https://via.placeholder.com/150",
      title: "Title 1",
      date: "May 22, 2024",
      description:
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas nec est sit amet libero efficitur tempor.",
    },
    {
      id: 2,
      image: "https://via.placeholder.com/150",
      title: "Title 2",
      date: "May 21, 2024",
      description:
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas nec est sit amet libero efficitur tempor.",
    },
    {
      id: 2,
      image: "https://via.placeholder.com/150",
      title: "Title 2",
      date: "May 21, 2024",
      description:
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas nec est sit amet libero efficitur tempor.",
    },
    {
      id: 2,
      image: "https://via.placeholder.com/150",
      title: "Title 2",
      date: "May 21, 2024",
      description:
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas nec est sit amet libero efficitur tempor.",
    },
    {
      id: 2,
      image: "https://via.placeholder.com/150",
      title: "Title 2",
      date: "May 21, 2024",
      description:
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas nec est sit amet libero efficitur tempor.",
    },
    {
      id: 2,
      image: "https://via.placeholder.com/150",
      title: "Title 2",
      date: "May 21, 2024",
      description:
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas nec est sit amet libero efficitur tempor.",
    },
    {
      id: 2,
      image: "https://via.placeholder.com/150",
      title: "Title 2",
      date: "May 21, 2024",
      description:
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas nec est sit amet libero efficitur tempor.",
    },
    {
      id: 2,
      image: "https://via.placeholder.com/150",
      title: "Title 2",
      date: "May 21, 2024",
      description:
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas nec est sit amet libero efficitur tempor.",
    },
    {
      id: 2,
      image: "https://via.placeholder.com/150",
      title: "Title 2",
      date: "May 21, 2024",
      description:
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas nec est sit amet libero efficitur tempor.",
    },
    {
      id: 2,
      image: "https://via.placeholder.com/150",
      title: "Title 2",
      date: "May 21, 2024",
      description:
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas nec est sit amet libero efficitur tempor.",
    },
    {
      id: 2,
      image: "https://via.placeholder.com/150",
      title: "Title 2",
      date: "May 21, 2024",
      description:
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas nec est sit amet libero efficitur tempor.",
    },
    {
      id: 2,
      image: "https://via.placeholder.com/150",
      title: "Title 2",
      date: "May 21, 2024",
      description:
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas nec est sit amet libero efficitur tempor.",
    },
    {
      id: 2,
      image: "https://via.placeholder.com/150",
      title: "Title 2",
      date: "May 21, 2024",
      description:
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas nec est sit amet libero efficitur tempor.",
    },
    {
      id: 2,
      image: "https://via.placeholder.com/150",
      title: "Title 2",
      date: "May 21, 2024",
      description:
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas nec est sit amet libero efficitur tempor.",
    },
    {
      id: 2,
      image: "https://via.placeholder.com/150",
      title: "Title 2",
      date: "May 21, 2024",
      description:
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas nec est sit amet libero efficitur tempor.",
    },
  ];

  return (
    <div className="news-container">
      {newsItems.map((item) => (
        <div key={item.id} className="news-item">
          <img src={item.image} alt={item.title} className="news-image" />
          <div className="news-content">
            <h3 className="news-title">{item.title}</h3>
            <p className="news-date">{item.date}</p>
            <p className="news-description">
              {item.description?.slice(0, 40)?.trim()}.....
            </p>
          </div>
        </div>
      ))}
    </div>
  );
}

export default NewsAndUpdates;
