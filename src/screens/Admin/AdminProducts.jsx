import { faPlus } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { useState, useEffect } from "react";
import {
  Modal,
  Button,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
} from "@mui/material";

import { Table } from "reactstrap"; // Import Reactstrap components
import { Categories } from "../../data/categories";
import { auth, db, storage } from "../../firebase";
import ShowAlert from "../../utils/showAlert";

// Sample product data (replace with actual data from Firebase)

function AdminProducts() {
  const [products, setProducts] = useState([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState(null);
  const [productName, setProductName] = useState("");
  const [productPrice, setProductPrice] = useState(0);
  const [productCategory, setProductCategory] = useState("");
  const [productImages, setProductImages] = useState([]);
  const [productQuantity, setProductQuantity] = useState();
  const [imagePreview, setImagePreview] = useState([]);
  const [updatingProduct, setUpdatingProduct] = useState(false);

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
      });
  }, []);

  const toggleModal = () => {
    setModalOpen(!modalOpen);
    setSelectedProduct(null); // Reset selected product when the modal is closed
    setProductName("");
    setProductPrice(0);
    setProductCategory("");
    setProductImages([]);
    setImagePreview([]);
    setProductQuantity(0);
  };

  const openEditModal = (product) => {
    setSelectedProduct(product);
    setModalOpen(true);
    setProductName(product.name);
    setProductPrice(product.price);
    setProductCategory(product.category);
    setProductImages(product.images);
    setImagePreview(product.images);
    setProductQuantity(product.quantity);
  };

  const uploadImagesToStorage = async () => {
    const imageUrls = [];

    try {
      for (const image of productImages) {
        const imageRef = storage.ref().child(`productImages/${image.name}`);
        await imageRef.put(image);
        const imageUrl = await imageRef.getDownloadURL();
        imageUrls.push(imageUrl);
      }
    } catch (error) {
      console.error("Error uploading images:", error);
    }

    return imageUrls;
  };

  const handleImageChange = (e) => {
    const selectedFiles = Array.from(e.target.files);

    // Display image preview
    const previews = selectedFiles.map((file) => URL.createObjectURL(file));
    setImagePreview(previews);

    // Store selected images
    setProductImages(selectedFiles);
  };
  const saveProduct = async () => {
    try {
      console.log("Saving product...");

      let imageUrls = [];

      // Upload images and get image URLs
      if (!updatingProduct) {
        imageUrls = await uploadImagesToStorage();
      } else {
        imageUrls = productImages;
      }

      // Create a new product object with updated data, including image URLs
      const updatedProduct = {
        name: productName,
        price: productPrice,
        collection: productCategory,
        quantity: productQuantity,
        images: imageUrls, // Use the uploaded image URLs
        owner: auth.currentUser.displayName,
        businessName: "TODO - Replace with dynamic business name",
      };

      if (updatingProduct) {
        // Update the product in the database
        await db
          .collection("Products")
          .doc(selectedProduct.id)
          .update(updatedProduct);

        const index = products.findIndex(
          (product) => product.id === selectedProduct.id
        );
        const updatedProducts = [...products];
        updatedProducts[index] = updatedProduct;
        setProducts(updatedProducts);
      } else {
        // Add the product to the database
        await db.collection("Products").doc().set(updatedProduct);
        // Add the product to the products state
        setProducts((prevProducts) => [...prevProducts, updatedProduct]);
      }

      // Close the modal
      setUpdatingProduct(false);
      toggleModal();
    } catch (error) {
      console.error("Error saving product:", error);
    }
  };

  const deleteProduct = (productId) => {
    db.collection("Products")
      .doc(productId)
      .delete()
      .then(() => {
        ShowAlert("Product deleted successfully", "success");
      })
      .catch((error) => {
        console.error("Error deleting product:", error);
      });

    // Remove the product from the products state
    const updatedProducts = products.filter(
      (product) => product.id !== productId
    );
    setProducts(updatedProducts);
  };

  return (
    <div className={"adminProducts"}>
      <div className={"header-admin"}>
        <h1>Product Management</h1>
        <Button variant="contained" color="primary" onClick={toggleModal}>
          New{" "}
          <FontAwesomeIcon
            icon={faPlus}
            size="1x"
            style={{
              marginLeft: "5px",
            }}
          />
        </Button>
      </div>

      <Table className={"table"}>
        <thead>
          <tr>
            <th>Image</th>
            <th>Name</th>
            <th>Collection</th>
            <th>Quantity</th>
            <th>Price</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {products.map((product) => (
            <tr key={product.id}>
              <td>
                <img
                  src={product.images[0]}
                  alt={product.name}
                  style={{
                    width: "100px",
                  }}
                />
              </td>
              <td>{product.name}</td>
              <td>{product.category}</td>
              <td>{product?.quantity}</td>
              <td>${product.price.toFixed(2)}</td>
              <td>
                <Button
                  variant="contained"
                  color="primary"
                  onClick={() => {
                    openEditModal(product);
                    setUpdatingProduct(true);
                  }}
                >
                  Edit
                </Button>{" "}
                <Button
                  variant="contained"
                  color="primary"
                  onClick={() => deleteProduct(product.id)}
                >
                  Delete
                </Button>
              </td>
            </tr>
          ))}
        </tbody>
      </Table>

      {/* Add/Edit Product Modal */}
      <Modal open={modalOpen} onClose={toggleModal}>
        <div className="modal-content">
          <h1>{selectedProduct ? "Edit Product" : "Add Product"}</h1>
          <div>
            <TextField
              label="Product Name"
              variant="outlined"
              fullWidth
              margin="normal"
              value={productName}
              onChange={(e) => setProductName(e.target.value)}
            />
            <TextField
              label="Price"
              variant="outlined"
              fullWidth
              margin="normal"
              type="number"
              value={productPrice === 0 ? "" : productPrice}
              onChange={(e) => setProductPrice(+e.target.value)}
            />
            <TextField
              label="Quantity"
              variant="outlined"
              fullWidth
              margin="normal"
              type="number"
              value={productQuantity === 0 ? "" : productQuantity}
              onChange={(e) => setProductQuantity(+e.target.value)}
            />
            <FormControl fullWidth margin="normal">
              <InputLabel>Category</InputLabel>
              <Select
                value={productCategory}
                onChange={(e) => setProductCategory(e.target.value)}
              >
                <MenuItem value="">
                  <em>Select a Collection</em>
                </MenuItem>
                {Categories.map((category) => (
                  <MenuItem key={category.title} value={category.title}>
                    {category.title}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <div>
              <label htmlFor="productImages">Images</label>
              <br />
              {!updatingProduct && (
                <Button
                  variant="contained"
                  color="primary"
                  onClick={() => {
                    document.getElementById("productImages").click();
                  }}
                >
                  Add Images
                </Button>
              )}
              <input
                type="file"
                name="productImages"
                id="productImages"
                multiple
                onChange={handleImageChange}
                style={{
                  display: "none",
                }}
              />
              {imagePreview.map((preview, index) => (
                <img
                  key={index}
                  src={preview}
                  alt={`Image ${index + 1}`}
                  className="image-preview"
                  width={50}
                  height={50}
                />
              ))}
            </div>
          </div>
          <div>
            <Button
              variant="contained"
              color="primary"
              onClick={async () => {
                await saveProduct(uploadImagesToStorage);
                toggleModal();
              }}
            >
              Save
            </Button>{" "}
            <Button variant="contained" color="secondary" onClick={toggleModal}>
              Cancel
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}

export default AdminProducts;
