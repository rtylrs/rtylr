import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

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
    // Add more product items here
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
          {products.map((product) => (
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
    </div>
  );
}

export default Home;
