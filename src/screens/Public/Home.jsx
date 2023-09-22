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
    </div>
  );
}

export default Home;
