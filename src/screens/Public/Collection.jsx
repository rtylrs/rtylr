import { useNavigate, useParams } from "react-router-dom";
import { faArrowLeft } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import ProductsGrid from "../../components/ProductsGrid";
import { useSelector } from "react-redux";
import { useEffect, useState } from "react";
import Loader from "../../layouts/loader/Loader";

function Collection() {
  const { collectionName } = useParams();
  const navigate = useNavigate();

  // Use useSelector to get products from Redux store
  const allProducts = useSelector((state) => state.products.products);

  const [products, setProducts] = useState([]);
  const [isLoading, setIsLoading] = useState(true);

  const fetchProductsByCollection = async () => {
    const filteredProducts = allProducts.filter((product) => {
      const isMatching =
        product.category.toLowerCase() === collectionName.toLowerCase();
      return isMatching;
    });

    setProducts(filteredProducts);
    setIsLoading(false);
  };

  useEffect(() => {
    fetchProductsByCollection();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleBack = () => {
    navigate(-1);
  };

  if (isLoading) {
    return <Loader />;
  }

  return (
    <div className="collection-container">
      <div className="header">
        <button onClick={handleBack}>
          <FontAwesomeIcon icon={faArrowLeft} />
        </button>
        <h1>{collectionName?.toUpperCase()}</h1>
      </div>
      <div className="productsGridContainer">
        <ProductsGrid products={products} />
      </div>
    </div>
  );
}

export default Collection;
