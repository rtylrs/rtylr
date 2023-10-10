import { useState, useEffect } from "react";
import {
  Button,
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
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faTrash } from "@fortawesome/free-solid-svg-icons";
import ShowAlert from "../../utils/showAlert";

function AdminContacts() {
  const [contacts, Contacts] = useState([]);

  useEffect(() => {
    db.collection("contacts")
      .get()
      .then((querySnapshot) => {
        const data = querySnapshot.docs.map((doc) => ({
          id: doc.id,
          ...doc.data(),
        }));
        Contacts(data);
      });
  }, []);

  return (
    <div className="AdminUsersManagement">
      <div className="AdminUsersManagementInner">
        <Typography variant="h3">Messages</Typography>
        <TableContainer component={Paper}>
          <Table className="adminTable">
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Email</TableCell>
                <TableCell>Phone</TableCell>
                <TableCell>Message</TableCell>
                <TableCell>Read</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {contacts &&
                contacts.map((contacts) => (
                  <TableRow key={contacts.id}>
                    <TableCell>{contacts.name}</TableCell>
                    <TableCell>{contacts.email}</TableCell>
                    <TableCell>{contacts.phone}</TableCell>
                    <TableCell>{contacts.message}</TableCell>
                    <TableCell>
                      <Button
                        variant="contained"
                        color="primary"
                        onClick={() => {
                          db.collection("contacts")
                            .doc(contacts.id)
                            .delete()
                            .then(() => {
                              const newContacts = contacts.filter(
                                (contact) => contact.id !== contacts.id
                              );
                              Contacts(newContacts);
                              ShowAlert("Message Read successfully", "success");
                            })
                            .catch((error) => {
                              console.error("Error removing document: ", error);
                            });
                        }}
                      >
                        <FontAwesomeIcon icon={faTrash} />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
            </TableBody>
          </Table>
        </TableContainer>
      </div>
    </div>
  );
}

export default AdminContacts;
