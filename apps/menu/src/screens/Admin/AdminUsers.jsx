import { useEffect, useState } from "react";
import { CSVLink } from "react-csv";
import { Button, Input } from "@voxire/components";
import { FaCheck, FaTrash, FaTimes } from "react-icons/fa";

function Users() {
  const [items, setItems] = useState([]);
  const [filteredItems, setFilteredItems] = useState([]);

  useEffect(() => {}, []);

  const handleSearch = (e) => {
    const searchValue = e.target.value.toLowerCase();
    setFilteredItems(
      items.filter((item) => item.name.toLowerCase().includes(searchValue))
    );
  };

  return (
    <div className="AdminUsersManagement">
      <div className="AdminUsersManagementInner">
        <h3>Users</h3>
        <div id="vendorManAdmin">
          <div>
            <Input
              type="text"
              name="search"
              id="search"
              placeholder="Search by name"
              onChange={handleSearch}
            />
          </div>
          <div className="exportBtnVen">
            <CSVLink data={items} filename={"vendors.csv"}>
              Export to CSV
            </CSVLink>
          </div>
        </div>
        <div className="ResponsiveTable">
          <div className="table-container">
            <table className="UsersTable">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Name</th>
                  <th>Email</th>
                  <th>Number of Purchased Items</th>
                  <th>Flagged</th>
                  <th>Delete</th>
                </tr>
              </thead>
              <tbody>
                {filteredItems.map((item, index) => (
                  <tr key={index}>
                    <td>{item.id}</td>
                    <td>{item.name}</td>
                    <td>{item.email}</td>
                    <td>{item.NumberOfEvents ? item.NumberOfEvents : 0}</td>
                    <td>
                      <Button variant="primary" onClick={() => {}}>
                        {item.flagged ? <FaCheck /> : <FaTimes />}
                      </Button>
                    </td>
                    <td>
                      <Button variant="primary" onClick={() => {}}>
                        <FaTrash />
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Users;
