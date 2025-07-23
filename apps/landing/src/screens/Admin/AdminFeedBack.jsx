import { useState, useEffect } from "react";
import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from "@mui/material";

function AdminFeedBack() {
  const [feedbacks, setFeedbacks] = useState([]);

  useEffect(() => {}, []);

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
