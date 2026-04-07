import React, { useState, useEffect } from "react";
import { Button, Modal, Input } from "@voxire/components";

function AdminBlog() {
  const [blogs, setBlogs] = useState([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [newBlog, setNewBlog] = useState({ title: "", content: "" });
  const [editedBlog, setEditedBlog] = useState({ id: null, title: "", content: "" });

  useEffect(() => {}, []);

  const handleOpenModal = () => setIsModalOpen(true);
  const handleCloseModal = () => { setIsModalOpen(false); setNewBlog({ title: "", content: "" }); };
  const handleOpenEditModal = (blog) => { setEditedBlog({ id: blog.id, title: blog.title, content: blog.content }); setIsEditModalOpen(true); };
  const handleCloseEditModal = () => { setIsEditModalOpen(false); setEditedBlog({ id: null, title: "", content: "" }); };

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    if (isEditModalOpen) {
      setEditedBlog((prev) => ({ ...prev, [name]: value }));
    } else {
      setNewBlog((prev) => ({ ...prev, [name]: value }));
    }
  };

  const handleAddBlog = () => {};
  const handleUpdateBlog = () => {};

  return (
    <div style={{ padding: "1rem" }}>
      <h1>Blogs</h1>
      <div className="table-container">
        <table className="adminTable">
          <thead>
            <tr>
              <th>Title</th>
              <th>Content</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {blogs && blogs.map((blog) => (
              <tr key={blog.id}>
                <td>{blog.title}</td>
                <td>{blog.content}</td>
                <td>
                  <Button variant="primary" onClick={() => handleOpenEditModal(blog)}>Edit</Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <Button variant="primary" onClick={handleOpenModal} style={{ marginTop: 16 }}>
        Add New Blog
      </Button>

      <Modal
        open={isModalOpen || isEditModalOpen}
        onClose={isEditModalOpen ? handleCloseEditModal : handleCloseModal}
        title={isEditModalOpen ? "Edit Blog" : "Add New Blog"}
        maxWidth="sm"
      >
        <div className="modal-form">
          <Input
            name="title"
            label="Title"
            hasLabel
            value={isEditModalOpen ? editedBlog.title : newBlog.title}
            onChange={handleInputChange}
          />
          <div style={{ marginTop: 12 }}>
            <label>Content</label>
            <textarea
              name="content"
              rows={4}
              style={{ width: "100%", padding: "8px", marginTop: 4, borderRadius: 4, border: "1px solid #d1d5db" }}
              value={isEditModalOpen ? editedBlog.content : newBlog.content}
              onChange={handleInputChange}
            />
          </div>
          <Button
            variant="primary"
            onClick={isEditModalOpen ? handleUpdateBlog : handleAddBlog}
            style={{ marginTop: 16 }}
          >
            {isEditModalOpen ? "Update Blog" : "Add Blog"}
          </Button>
        </div>
      </Modal>
    </div>
  );
}

export default AdminBlog;
