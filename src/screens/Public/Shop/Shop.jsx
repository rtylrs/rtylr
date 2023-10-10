import { useState, useEffect } from "react";
import CategorySlider from "../../../components/CategorySlider";
import ProductCard from "../../../components/ProductCard";
import { useSelector } from "react-redux";

function Shop() {
  const allProducts = useSelector((state) => state.products.products);
  const [paginatedProducts, setPaginatedProducts] = useState([]);
  const [page, setPage] = useState(1); // Track the current page

  // Load initial products from allProducts
  useEffect(() => {
    setPaginatedProducts(allProducts.slice(0, 9));
  }, [allProducts]);

  const loadMoreProducts = () => {
    const startIndex = page * 9;
    const endIndex = startIndex + 9;

    // Check if there are more products to load
    if (startIndex < allProducts.length) {
      const newProducts = allProducts.slice(startIndex, endIndex);
      setPaginatedProducts((prevProducts) => [...prevProducts, ...newProducts]);
      setPage(page + 1);
    }
  };

  return (
    <div>
      <CategorySlider />
      <div className="shop">
        {/* Filter section */}
        <aside className="filter-sidebar">
          <h3>Filters</h3>
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
          {paginatedProducts &&
            paginatedProducts.map((product) => (
              <ProductCard key={product.id} product={product} />
            ))}
        </div>
      </div>
      <div className="shopFooter">
        <button className="view-more-button-shop" onClick={loadMoreProducts}>
          Load More
        </button>
      </div>
    </div>
  );
}

export default Shop;
