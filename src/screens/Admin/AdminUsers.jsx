import { useEffect, useState } from "react";
import { db } from "../../firebase";
import { CSVLink } from "react-csv";
import { LazyLoadImage } from "react-lazy-load-image-component";
import "react-lazy-load-image-component/src/effects/blur.css";
import Button from "@mui/material/Button";
import { Table } from "reactstrap";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faTrash } from "@fortawesome/free-solid-svg-icons";

function Users() {
  const [items, setItems] = useState([]);

  useEffect(() => {
    db.collection("Users")
      .limit(15)
      .onSnapshot((snapshot) => {
        if (!snapshot.empty) {
          let item = [];
          snapshot.forEach((doc) => {
            let data = Object.assign({ id: doc.id }, doc.data());
            item.push(data);
          });
          setItems(item);
        }
      });
  }, []);

  return (
    <div className="AdminUsersManagement">
      <div className="AdminUsersManagementInner">
        <h3>Users</h3>
        <div id="vendorManAdmin">
          <div>
            <input
              type="text"
              name="search"
              id="search"
              placeholder="Search by name"
              onChange={(e) => {
                setItems(
                  items.filter((item) =>
                    item.name.toLowerCase().includes(e.target.value)
                  )
                );
              }}
            />
          </div>
          <div className="exportBtnVen">
            <CSVLink data={items} filename={"vendors.csv"}>
              Export to CSV
            </CSVLink>
          </div>
        </div>
        <div className="ResponsiveTable">
          <Table className="UsersTable">
            <thead>
              <tr>
                <th>Name</th>
                <th>Email</th>
                <th>Number Of Logins</th>
                <th>Number of Purchased Items</th>
                <th>Delete</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item, index) => {
                return (
                  <tr key={index}>
                    <td>{item.name}</td>
                    <td>{item.email}</td>
                    <td>{item.NumberOfLogins}</td>
                    <td>{item.NumberOfEvents ? item.NumberOfEvents : 0}</td>
                    <td>
                      <Button
                        variant="contained"
                        color="primary"
                        onClick={() => {
                          // do something
                        }}
                      >
                        <FontAwesomeIcon icon={faTrash} size="lg" />
                      </Button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </Table>
        </div>
      </div>
    </div>
  );
}

export default Users;
