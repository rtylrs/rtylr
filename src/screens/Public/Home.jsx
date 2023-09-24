import { useTranslation } from "react-i18next";
import { Link, useNavigate } from "react-router-dom";
import { Categories } from "../../data/categories";
import { useEffect, useState } from "react";
import { db } from "../../firebase";
import { LazyLoadImage } from "react-lazy-load-image-component";

function Home() {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [blogs, setBlogs] = useState([]);

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

  useEffect(() => {
    db.collection("Blogs")
      .limit(3)
      .get()
      .then((querySnapshot) => {
        const data = querySnapshot.docs.map((doc) => ({
          id: doc.id,
          ...doc.data(),
        }));
        setBlogs(data);
      });
  }, []);

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
                <LazyLoadImage
                  src={product.imageUrl}
                  alt={`Product ${product.id}`}
                  effect="blur"
                  width="100%"
                  height="230px"
                />
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
          {blogs &&
            blogs.map((post, index) => (
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
                <p className="post-date">
                  {post.timestamp.toDate().toDateString()}
                </p>
              </Link>
            ))}
        </div>
      </section>
    </div>
  );
}

export default Home;
