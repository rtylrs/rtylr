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

// Sample product data (replace with actual data from Firebase)
const initialProducts = [
  {
    id: "1",
    category: "Tote Bags",
    collection: "bags",
    creator: "Kodra",
    description: ["test 1", "test 2"],
    images: [
      "https://example.com/image1.jpg",
      "https://example.com/image2.jpg",
    ],
    name: "Test Product 1",
    price: 25,
  },
  {
    id: "2",
    category: "Backpacks",
    collection: "bags",
    creator: "John Doe",
    description: ["Description 1"],
    images: ["https://example.com/image3.jpg"],
    name: "Test Product 2",
    price: 30,
  },
];

function AdminProducts() {
  const [products, setProducts] = useState(initialProducts);
  const [modalOpen, setModalOpen] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState(null);
  const [productName, setProductName] = useState("");
  const [productPrice, setProductPrice] = useState("");
  const [productCategory, setProductCategory] = useState("");
  const [productImages, setProductImages] = useState([]);
  const [imagePreview, setImagePreview] = useState([]);

  const toggleModal = () => {
    setModalOpen(!modalOpen);
    setSelectedProduct(null); // Reset selected product when the modal is closed
    setProductName("");
    setProductPrice("");
    setProductCategory("");
    setProductImages([]);
    setImagePreview([]);
  };

  const openEditModal = (product) => {
    setSelectedProduct(product);
    setModalOpen(true);
    setProductName(product.name);
    setProductPrice(product.price);
    setProductCategory(product.category);
    setProductImages(product.images);
    setImagePreview(product.images);
  };

  const handleImageChange = (e) => {
    const selectedFiles = Array.from(e.target.files);

    // Display image preview
    const previews = selectedFiles.map((file) => URL.createObjectURL(file));
    setImagePreview(previews);

    // Store selected images
    setProductImages(selectedFiles);
  };
  const saveProduct = () => {
    // Create a new product object with updated data
    const updatedProduct = {
      id: selectedProduct ? selectedProduct.id : Date.now().toString(),
      name: productName,
      price: productPrice,
      category: productCategory,
      images: productImages,
    };

    // Implement saving or updating product to Firebase here
    // After saving, update the products state
    // For now, just log the data to the console
    console.log("Saved Product:", updatedProduct);

    // Close the modal
    toggleModal();
  };

  const deleteProduct = (productId) => {
    // Implement product deletion logic here
    // After deleting, update the products state
    // For now, just log the product ID to the console
    console.log("Deleted Product ID:", productId);
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
            <th>Category</th>
            <th>Collection</th>
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
              <td>{product.collection}</td>
              <td>${product.price.toFixed(2)}</td>
              <td>
                <Button
                  variant="contained"
                  color="primary"
                  onClick={() => openEditModal(product)}
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
                onChange={(e) => setProductPrice(e.target.value)}
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
                <option value="">Select a category</option>
                {Categories.map((category) => (
                  <option key={category.title} value={category.title}>
                    {category.title}
                  </option>
                ))}
              </Input>
            </FormGroup>
            <FormGroup>
              <Label for="productImages">Images</Label>
              <Input
                type="file"
                name="productImages"
                id="productImages"
                multiple
                onChange={handleImageChange}
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
