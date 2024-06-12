import { useEffect, useState } from "react";
import { CSVLink } from "react-csv";
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
  Button,
  Paper,
} from "@mui/material";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCheck, faTrash, faX } from "@fortawesome/free-solid-svg-icons";

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
        <Typography variant="h3">Users</Typography>
        <div id="vendorManAdmin">
          <div>
            <TextField
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
          <TableContainer component={Paper}>
            <Table className="UsersTable">
              <TableHead>
                <TableRow>
                  <TableCell>ID</TableCell>
                  <TableCell>Name</TableCell>
                  <TableCell>Email</TableCell>
                  <TableCell>Number of Purchased Items</TableCell>
                  <TableCell>Flagged</TableCell>
                  <TableCell>Delete</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {filteredItems.map((item, index) => (
                  <TableRow key={index}>
                    <TableCell>{item.id}</TableCell>
                    <TableCell>{item.name}</TableCell>
                    <TableCell>{item.email}</TableCell>
                    <TableCell>
                      {item.NumberOfEvents ? item.NumberOfEvents : 0}
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="contained"
                        color="primary"
                        onClick={() => {
                          // Handle delete action
                        }}
                      >
                        {item.flagged ? (
                          <FontAwesomeIcon icon={faCheck} size="lg" />
                        ) : (
                          <FontAwesomeIcon icon={faX} size="lg" />
                        )}
                      </Button>
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="contained"
                        color="primary"
                        onClick={() => {
                          // Handle delete action
                        }}
                      >
                        <FontAwesomeIcon icon={faTrash} size="lg" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </div>
      </div>
    </div>
  );
}

export default Users;
