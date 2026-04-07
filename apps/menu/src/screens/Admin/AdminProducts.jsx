import { useState, useEffect } from "react";
import { Button, Modal, Input } from "@voxire/components";
import { FaPlus } from "react-icons/fa";

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

  useEffect(() => {}, []);

  const toggleModal = () => {
    setModalOpen(!modalOpen);
    setSelectedProduct(null);
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

  const uploadImagesToStorage = async () => [];
  const saveProduct = async () => {};
  const deleteProduct = (productId) => {};

  const handleImageChange = (e) => {
    const selectedFiles = Array.from(e.target.files);
    const previews = selectedFiles.map((file) => URL.createObjectURL(file));
    setImagePreview(previews);
    setProductImages(selectedFiles);
  };

  return (
    <div className="AdminUsersManagement">
      <div className="AdminUsersManagementInner">
        <div className="header-admin">
          <h1>Product Management</h1>
          <Button variant="primary" onClick={toggleModal}>
            New <FaPlus style={{ marginLeft: 5 }} />
          </Button>
        </div>

        <div className="table-container">
          <table>
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
                  <td><img src={product.images[0]} alt={product.name} style={{ width: 100 }} /></td>
                  <td>{product.name}</td>
                  <td>{product.category}</td>
                  <td>{product.quantity}</td>
                  <td>${product.price.toFixed(2)}</td>
                  <td>
                    <Button variant="primary" onClick={() => { openEditModal(product); setUpdatingProduct(true); }}>
                      Edit
                    </Button>{" "}
                    <Button variant="primary" onClick={() => deleteProduct(product.id)}>
                      Delete
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <Modal open={modalOpen} onClose={toggleModal} title={selectedProduct ? "Edit Product" : "Add Product"} maxWidth="sm">
          <div className="modal-form">
            <Input label="Product Name" hasLabel value={productName} onChange={(e) => setProductName(e.target.value)} />
            <Input label="Price" hasLabel type="number" value={productPrice === 0 ? "" : productPrice} onChange={(e) => setProductPrice(+e.target.value)} />
            <Input label="Quantity" hasLabel type="number" value={productQuantity === 0 ? "" : productQuantity} onChange={(e) => setProductQuantity(+e.target.value)} />
            <div>
              <label htmlFor="category">Category</label>
              <select id="category" value={productCategory} onChange={(e) => setProductCategory(e.target.value)}>
                <option value="">Select a Collection</option>
              </select>
            </div>
            <div>
              <label>Images</label>
              {!updatingProduct && (
                <Button variant="primary" onClick={() => document.getElementById("productImages").click()}>
                  Add Images
                </Button>
              )}
              <input type="file" id="productImages" multiple onChange={handleImageChange} style={{ display: "none" }} />
              {imagePreview.map((preview, index) => (
                <img key={index} src={preview} alt={`Image ${index + 1}`} className="image-preview" width={50} height={50} />
              ))}
            </div>
            <div style={{ display: "flex", gap: 8, marginTop: 16 }}>
              <Button variant="primary" onClick={async () => { await saveProduct(uploadImagesToStorage); toggleModal(); }}>
                Save
              </Button>
              <Button variant="outline" onClick={toggleModal}>Cancel</Button>
            </div>
          </div>
        </Modal>
      </div>
    </div>
  );
}

export default AdminProducts;
