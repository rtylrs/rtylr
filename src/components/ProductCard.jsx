import PropTypes from "prop-types";
import { Link, useNavigate } from "react-router-dom";
import { LazyLoadImage } from "react-lazy-load-image-component";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faShoppingCart } from "@fortawesome/free-solid-svg-icons";

function ProductCard({ product }) {
  const navigate = useNavigate();
  return (
    <div>
      <div className="productCard">
        <Link className="imageContainer" to={`/shop/${product?.id}`}>
          <LazyLoadImage
            src={product?.images[0]}
            alt="Product image"
            effect="blur"
          />
        </Link>
        <h4 className="name">{product?.name}</h4>
        <p className="creator">
          By <strong>{product?.creator}</strong> with &#9829;
        </p>
        <p className="price">{product?.price} KD</p>
        <div className="ctaContainer">
          <button className="addBtn">
            <span>
              <FontAwesomeIcon icon={faShoppingCart} />
            </span>
            <span>{product?.inCart ? "Remove" : "Add"}</span>
          </button>
          <button className="buyBtn" onClick={() => {}}>
            Buy Now
          </button>
        </div>
      </div>
    </div>
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
    category: PropTypes.string.isRequired,
    inCart: PropTypes.number,
  }).isRequired,
};

export default ProductCard;
