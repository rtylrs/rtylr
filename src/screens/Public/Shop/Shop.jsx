import { useState, useEffect } from "react";
import CategorySlider from "../../../components/CategorySlider";
import ProductCard from "../../../components/ProductCard";
import { db } from "../../../firebase";

function Shop() {
  const [products, setProducts] = useState([]);
  const [lastProductLoaded, setLastProductLoaded] = useState(0);

  useEffect(() => {
    db.collection("Products")
      .limit(8)
      .get()
      .then((querySnapshot) => {
        const products = [];
        querySnapshot.forEach((doc) => {
          const product = Object.assign({ id: doc.id }, doc.data());
          products.push(product);
        });
        setProducts(products);

        // Set the last product loaded
        setLastProductLoaded(products[products.length - 1]);
      });
  }, []);

  const loadMoreProducts = () => {
    db.collection("products")
      .startAfter(lastProductLoaded)
      .limit(8)
      .get()
      .then((querySnapshot) => {
        const products = [];
        querySnapshot.forEach((doc) => {
          products.push(doc.data());
        });
        setProducts((prevProducts) => [...prevProducts, ...products]);

        // Set the last product loaded
        setLastProductLoaded(products[products.length - 1]);
      });
  };

  return (
    <div>
      <CategorySlider />
      <div className="shop">
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
          {products &&
            products.map((product) => (
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
