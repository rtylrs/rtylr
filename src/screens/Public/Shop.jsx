import { useState } from "react";
import CategorySlider from "../../components/CategorySlider";
import ProductCard from "../../components/ProductCard";

const initialProducts = [
  {
    id: 1,
    name: "Product 1",
    price: 25,
    creator: "Creator A",
    images: ["product1.jpg", "product1-2.jpg"],
  },
  {
    id: 2,
    name: "Product 2",
    price: 30,
    creator: "Creator B",
    images: ["product2.jpg", "product2-2.jpg"],
  },
  // Add more products as needed...
];

function Shop() {
  const [products, setProducts] = useState(initialProducts);

  const loadMoreProducts = () => {
    // You can fetch more products from an API or add more to the state here.
  };

  return (
    <div>
      <CategorySlider />
      <div className="shop">
        <aside className="filter-sidebar">
          <h3>Filter</h3>
          <div className="filter-section">
            <h4>Price</h4>
            <input type="range" min="0" max="100" />
          </div>
          <div className="filter-section">
            <h4>Creator</h4>
            <input type="checkbox" />
            <label>Creator A</label>
            <br />
            <input type="checkbox" />
            <label>Creator B</label>
            <br />
            <input type="checkbox" />
            <label>Creator C</label>
          </div>
        </aside>
        <div className="product-grid">
          {products.map((product) => (
            <ProductCard key={product.id} product={product} />
          ))}
        </div>
      </div>
      <button className="view-more-button" onClick={loadMoreProducts}>
        View More
      </button>
    </div>
  );
}

export default Shop;
