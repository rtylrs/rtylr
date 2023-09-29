import { useNavigate, useParams } from "react-router-dom";
import { faArrowLeft } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import ProductsGrid from "../../components/ProductsGrid";
import { db } from "../../firebase";
import { useEffect, useState } from "react";

function Collection() {
  const { collectionName } = useParams();
  const navigate = useNavigate();

  const [products, setProducts] = useState([]);
  const [isLoading, setIsLoading] = useState(true);

  const fetchProductsByCollection = async () => {
    db.collection("Products")
      .where("collection", "==", collectionName)
      .get()
      .then((querySnapshot) => {
        const data = querySnapshot.docs.map((doc) => ({
          id: doc.id,
          ...doc.data(),
        }));
        setProducts(data);
        setIsLoading(false);
        console.log("products by collection", data);
      })
      .catch((error) => {
        console.log("Error getting documents: ", error);
      });
  };

  useEffect(() => {
    fetchProductsByCollection();
  }, []);

  const handleBack = () => {
    navigate(-1);
  };

  if (isLoading) {
    return (
      <div
        style={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          height: "100vh",
          width: "100%",
          fontSize: "2rem",
        }}
      >
        Loading...
      </div>
    );
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
