import React from "react";
import {
  Modal,
  Button,
  TextField,
  TableContainer,
  Table,
  TableHead,
  TableBody,
  TableRow,
  TableCell,
  Paper,
} from "@mui/material";
import { useState, useEffect } from "react";
import { Container } from "reactstrap";

function AdminBlog() {
  const [blogs, setBlogs] = useState([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [newBlog, setNewBlog] = useState({ title: "", content: "" });
  const [editedBlog, setEditedBlog] = useState({
    id: null,
    title: "",
    content: "",
  });

  useEffect(() => {}, []);

  const handleOpenModal = () => {
    setIsModalOpen(true);
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
    setNewBlog({ title: "", content: "" });
  };

  const handleOpenEditModal = (blog) => {
    setEditedBlog({ id: blog.id, title: blog.title, content: blog.content });
    setIsEditModalOpen(true);
  };

  const handleCloseEditModal = () => {
    setIsEditModalOpen(false);
    setEditedBlog({ id: null, title: "", content: "" });
  };

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    if (isEditModalOpen) {
      setEditedBlog((prevBlog) => ({
        ...prevBlog,
        [name]: value,
      }));
    } else {
      setNewBlog((prevBlog) => ({
        ...prevBlog,
        [name]: value,
      }));
    }
  };

  const handleAddBlog = () => {};

  const handleUpdateBlog = () => {};

  return (
    <Container>
      <h1>Blogs</h1>
      <TableContainer component={Paper}>
        <Table className="adminTable">
          <TableHead>
            <TableRow>
              <TableCell>Title</TableCell>
              <TableCell>Content</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {blogs &&
              blogs.map((blog) => (
                <TableRow key={blog.id}>
                  <TableCell>{blog.title}</TableCell>
                  <TableCell>{blog.content}</TableCell>
                  <TableCell>
                    <Button
                      variant="contained"
                      color="primary"
                      onClick={() => handleOpenEditModal(blog)}
                    >
                      Edit
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
          </TableBody>
        </Table>
      </TableContainer>
      <Button variant="contained" color="primary" onClick={handleOpenModal}>
        Add New Blog
      </Button>
      <Modal
        open={isModalOpen || isEditModalOpen}
        onClose={isEditModalOpen ? handleCloseEditModal : handleCloseModal}
      >
        <div className="modal-content">
          {isEditModalOpen ? (
            <>
              <h2>Edit Blog</h2>
              <TextField
                name="title"
                label="Title"
                variant="outlined"
                fullWidth
                margin="normal"
                value={editedBlog.title}
                onChange={handleInputChange}
              />
              <TextField
                name="content"
                label="Content"
                variant="outlined"
                fullWidth
                multiline
                rows={4}
                margin="normal"
                value={editedBlog.content}
                onChange={handleInputChange}
              />
              <Button
                variant="contained"
                color="primary"
                onClick={handleUpdateBlog}
              >
                Update Blog
              </Button>
            </>
          ) : (
            <>
              <h2>Add New Blog</h2>
              <TextField
                name="title"
                label="Title"
                variant="outlined"
                fullWidth
                margin="normal"
                value={newBlog.title}
                onChange={handleInputChange}
              />
              <TextField
                name="content"
                label="Content"
                variant="outlined"
                fullWidth
                multiline
                rows={4}
                margin="normal"
                value={newBlog.content}
                onChange={handleInputChange}
              />
              <Button
                variant="contained"
                color="primary"
                onClick={handleAddBlog}
              >
                Add Blog
              </Button>
            </>
          )}
        </div>
      </Modal>
    </Container>
  );
}

export default AdminBlog;
