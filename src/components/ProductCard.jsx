import PropTypes from "prop-types";
import { Link } from "react-router-dom";

function ProductCard({ product }) {
  return (
    <Link to={`/product/${product?.id}`}>
      <div className="productCard">
        <img src={product?.images[0]} alt="Product image" />
        <h3 className="name">Product 1</h3>
        <p className="creator">{product?.creator}</p>
        <p className="price">{product?.price} KD</p>
        <div className="ctaContainer">
          <button className="addBtn">
            <span>+</span>
            <span>Add</span>
          </button>
          <button className="buyBtn">Buy Now</button>
        </div>
      </div>
    </Link>
  );
}

/*id: PropTypes.number.isRequired,
      name: PropTypes.string.isRequired,
      price: PropTypes.number.isRequired,
      image: PropTypes.string.isRequired,
      collection: PropTypes.string.isRequired,*/

ProductCard.propTypes = {
  product: PropTypes.shape({
    id: PropTypes.string.isRequired,
    name: PropTypes.string.isRequired,
    price: PropTypes.number.isRequired,
    images: PropTypes.arrayOf(PropTypes.string).isRequired,
    creator: PropTypes.string.isRequired,
    collection: PropTypes.string.isRequired,
  }).isRequired,
};

export default ProductCard;
