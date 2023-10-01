import { faPlus } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { Button, Modal } from "@mui/material";
import { useState, useEffect } from "react";
import {
  Table,
  ModalHeader,
  ModalBody,
  ModalFooter,
  FormGroup,
  Label,
  Input,
} from "reactstrap"; // Import Reactstrap components
import { Categories } from "../../data/categories";
import { db, storage } from "../../firebase";
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
  const [productQuantity, setProductQuantity] = useState(0);
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
      };

      if (updatingProduct) {
        // Update the product in the database
        await db
          .collection("Products")
          .doc(selectedProduct.id)
          .update(updatedProduct);
      } else {
        // Add the product to the database
        await db.collection("Products").doc().set(updatedProduct);
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
          <ModalHeader toggle={toggleModal}>
            {selectedProduct ? "Edit Product" : "Add Product"}
          </ModalHeader>
          <ModalBody>
            <FormGroup>
              <Label for="productName">Product Name</Label>
              <Input
                type="text"
                name="productName"
                id="productName"
                value={productName}
                onChange={(e) => setProductName(e.target.value)}
              />
            </FormGroup>
            <FormGroup>
              <Label for="productPrice">Price</Label>
              <Input
                type="number"
                name="productPrice"
                id="productPrice"
                value={productPrice}
                onChange={(e) => {
                  setProductPrice(+e.target.value);
                }}
              />
            </FormGroup>
            <FormGroup>
              <Label for="quantity">Quatity</Label>
              <Input
                type="number"
                name="quantity"
                id="productQuantity"
                value={productQuantity}
                onChange={(e) => {
                  setProductQuantity(+e.target.value);
                }}
              />
            </FormGroup>
            <FormGroup>
              <Label for="productCategory">Category</Label>
              <Input
                type="select"
                name="productCategory"
                id="productCategory"
                value={productCategory}
                onChange={(e) => setProductCategory(e.target.value)}
              >
                <option value="">Select a Collection</option>
                {Categories.map((category) => (
                  <option key={category.title} value={category.title}>
                    {category.title}
                  </option>
                ))}
              </Input>
            </FormGroup>
            <FormGroup>
              <Label for="productImages">Images</Label>
              <br />
              {
                // Only show the "Add Images" button if the user is adding a new product
                !updatingProduct && (
                  <Button
                    variant="contained"
                    color="primary"
                    onClick={() => {
                      document.getElementById("productImages").click();
                    }}
                  >
                    Add Images
                  </Button>
                )
              }
              <Input
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
            </FormGroup>
          </ModalBody>
          <ModalFooter>
            <Button color="primary" onClick={saveProduct}>
              Save
            </Button>{" "}
            <Button color="secondary" onClick={toggleModal}>
              Cancel
            </Button>
          </ModalFooter>
        </div>
      </Modal>
    </div>
  );
}

export default AdminProducts;
