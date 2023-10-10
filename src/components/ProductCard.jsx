import PropTypes from "prop-types";
import { Link, useNavigate } from "react-router-dom";
import { auth } from "../firebase";
import { LazyLoadImage } from "react-lazy-load-image-component";

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
            <span>+</span>
            <span>Add</span>
          </button>
          <button
            className="buyBtn"
            onClick={() => {
              if (auth.currentUser) {
                console.log("user is signed in");
              } else {
                navigate("/auth/login");
              }
            }}
          >
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
  }).isRequired,
};

export default ProductCard;
