import PropTypes from "prop-types";
import ProductCard from "./ProductCard";

function ProductsGrid({ products }) {
  return (
    <div className="productsContainer">
      {products.map((product) => (
        <ProductCard key={product.id} product={product} />
      ))}
    </div>
  );
}

ProductsGrid.propTypes = {
  products: PropTypes.arrayOf(
    PropTypes.shape({
      id: PropTypes.string.isRequired,
      name: PropTypes.string.isRequired,
      price: PropTypes.number.isRequired,
      images: PropTypes.arrayOf(PropTypes.string).isRequired,
      creator: PropTypes.string.isRequired,
      category: PropTypes.string.isRequired,
      collection: PropTypes.string.isRequired,
      description: PropTypes.arrayOf(PropTypes.string).isRequired,
    })
  ).isRequired,
};

export default ProductsGrid;
