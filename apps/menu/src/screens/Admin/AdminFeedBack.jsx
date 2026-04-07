import { useState, useEffect } from "react";

function AdminFeedBack() {
  const [feedbacks, setFeedbacks] = useState([]);

  useEffect(() => {}, []);

  return (
    <div className="AdminUsersManagement">
      <div className="AdminUsersManagementInner">
        <h3>Feedback</h3>
        <div className="table-container">
          <table className="adminTable">
            <thead>
              <tr>
                <th>Name</th>
                <th>Email</th>
                <th>Rating</th>
              </tr>
            </thead>
            <tbody>
              {feedbacks && feedbacks.map((blog) => (
                <tr key={blog.id}>
                  <td>{blog.name}</td>
                  <td>{blog.email}</td>
                  <td>{blog.rating}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

export default AdminFeedBack;
