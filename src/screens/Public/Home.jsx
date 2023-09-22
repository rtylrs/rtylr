import { useTranslation } from "react-i18next";
import { Link, useNavigate } from "react-router-dom";
import { Categories } from "../../data/categories";

function Home() {
  const navigate = useNavigate();
  const { t } = useTranslation();

  const products = [
    {
      id: 1,
      name: "Product 1",
      price: "$10",
      businessName: "Business A",
      imageUrl: "https://via.placeholder.com/150", // Use a placeholder image from via.placeholder.com
    },
    {
      id: 2,
      name: "Product 2",
      price: "$15",
      businessName: "Business B",
      imageUrl: "https://via.placeholder.com/150", // Use another placeholder image
    },
    {
      id: 3,
      name: "Product 3",
      price: "$15",
      businessName: "Business B",
      imageUrl: "https://via.placeholder.com/150", // Use another placeholder image
    },
    {
      id: 4,
      name: "Product 4",
      price: "$15",
      businessName: "Business B",
      imageUrl: "https://via.placeholder.com/150", // Use another placeholder image
    },
  ];

  const blogPosts = [
    {
      id: 1,
      title: "Lorem Ipsum Blog Post 1",
      content: "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
      date: "September 1, 2023",
    },
    {
      id: 2,
      title: "Lorem Ipsum Blog Post 2",
      content:
        "Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
      date: "September 5, 2023",
    },
    {
      id: 3,
      title: "Lorem Ipsum Blog Post 3",
      content:
        "Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.",
      date: "September 10, 2023",
    },
  ];

  return (
    <div>
      <div className="Hero">
        <div className="hero-text">
          <h1>{t("Hero.welcome")}</h1>
          <h2>{t("Hero.discover")}</h2>
          <div className="searchBarHome">
            <input
              type="text"
              placeholder={t("Hero.searchPlaceholder")}
              className="searchBar"
            />
          </div>
        </div>
      </div>
      <section className="product-section">
        <h2 className="section-title">Browse Products</h2>
        <div className="product-grid">
          {products &&
            products.slice(0, 4).map((product) => (
              <div key={product.id} className="product-item">
                <img src={product.imageUrl} alt={`Product ${product.id}`} />
                <h3>{product.businessName}</h3>
                <p>{product.price}</p>
                <button>Buy Now</button>
              </div>
            ))}
        </div>
        <button
          className="view-more-button"
          onClick={() => {
            navigate("/shop");
          }}
        >
          View More
        </button>
      </section>

      <section className="home-section">
        <h2 className="section-title">Our Menu</h2>
        <div className="grid-container">
          {Categories.map((category, index) => (
            <Link
              key={index}
              className="grid-item"
              style={{
                backgroundImage: `url(${category.imageUrl})`,
              }}
              to={`/shop/${category.title.toLowerCase()}`}
            >
              <div className="item-content">
                <p>{category.title}</p>
              </div>
            </Link>
          ))}
        </div>
      </section>

      <section className="blog-section">
        <h2 className="section-title">Latest Blog Posts</h2>
        <div className="blog-grid">
          {blogPosts.map((post, index) => (
            <Link className="blog-post" key={index} to={`/blog/${post.id}`}>
              <h3 className="post-title">
                {post.title.length > 50
                  ? post.title.substring(0, 50) + "..."
                  : post.title}
              </h3>
              <p className="post-content">
                {post.content.length > 250
                  ? post.content.substring(0, 250) + "..."
                  : post.content}
              </p>
              <p className="post-date">{post.date}</p>
            </Link>
          ))}
        </div>
      </section>
    </div>
  );
}

export default Home;
