import React, { useState, useEffect } from "react";
import {
  Container,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from "@mui/material";
import { db } from "../../firebase";

function AdminFeedBack() {
  const [feedbacks, setFeedbacks] = useState([]);

  useEffect(() => {
    db.collection("feedback")
      .get()
      .then((querySnapshot) => {
        const data = querySnapshot.docs.map((doc) => ({
          id: doc.id,
          ...doc.data(),
        }));
        setFeedbacks(data);
      });
  }, []);

  return (
    <div className="AdminUsersManagement">
      <div className="AdminUsersManagementInner">
        <Typography variant="h3">Feedback</Typography>
        <TableContainer component={Paper}>
          <Table className="adminTable">
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Email</TableCell>
                <TableCell>Rating</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {feedbacks &&
                feedbacks.map((blog) => (
                  <TableRow key={blog.id}>
                    <TableCell>{blog.name}</TableCell>
                    <TableCell>{blog.email}</TableCell>
                    <TableCell>{blog.rating}</TableCell>
                  </TableRow>
                ))}
            </TableBody>
          </Table>
        </TableContainer>
      </div>
    </div>
  );
}

export default AdminFeedBack;
