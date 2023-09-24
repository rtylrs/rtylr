import { useState, useEffect } from "react";
import { Container, Table } from "reactstrap";
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
    <Container>
      <h1>Feedbacks</h1>
      <Table className="adminTable">
        <thead>
          <tr>
            <th>Name</th>
            <th>Email</th>
            <th>Rating</th>
          </tr>
        </thead>
        <tbody>
          {feedbacks &&
            feedbacks.map((blog) => (
              <tr key={blog.id}>
                <td>{blog.name}</td>
                <td>{blog.email}</td>
                <td>{blog.rating}</td>
              </tr>
            ))}
        </tbody>
      </Table>
    </Container>
  );
}

export default AdminFeedBack;
