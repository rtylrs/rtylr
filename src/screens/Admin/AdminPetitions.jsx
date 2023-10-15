import { useState, useEffect } from "react";
import {
  Button,
  Container,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import { db } from "../../firebase";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faTrash } from "@fortawesome/free-solid-svg-icons";

function AdminPetitions() {
  const [petitionTitle, setPetitionTitle] = useState("");
  const [petitionGoal, setPetitionGoal] = useState("");
  const [loading, setLoading] = useState(false);
  const [petitions, setPetitions] = useState([]);
  const [selectedUsers, setSelectedUsers] = useState([]);
  const [openModal, setOpenModal] = useState(false);

  const addPetition = (e) => {
    e.preventDefault();
    db.collection("petitions")
      .add({
        title: petitionTitle,
        goal: petitionGoal,
        signatures: 0,
        users: [],
      })
      .then(() => {
        setPetitionTitle("");
        setPetitionGoal("");
      });
  };

  useEffect(() => {
    setLoading(true);
    db.collection("petitions").onSnapshot((snapshot) => {
      setPetitions(
        snapshot.docs.map((doc) => ({
          id: doc.id,
          title: doc.data().title,
          goal: doc.data().goal,
          signatures: doc.data().signatures,
          users: doc.data().users,
        }))
      );
    });
    setLoading(false);
  }, []);

  const handleDelete = (id) => {
    db.collection("petitions").doc(id).delete();
  };

  const handleViewUsers = (users) => {
    setSelectedUsers(users);
    setOpenModal(true);
  };

  const handleCloseModal = () => {
    setOpenModal(false);
  };

  return (
    <Container>
      <form className="petitionsForm" onSubmit={addPetition}>
        <TextField
          label="Petition Title"
          variant="outlined"
          fullWidth
          value={petitionTitle}
          onChange={(e) => setPetitionTitle(e.target.value)}
        />
        <TextField
          label="Petition Goal"
          variant="outlined"
          fullWidth
          type="number"
          value={petitionGoal}
          onChange={(e) => setPetitionGoal(e.target.value)}
        />
        <Button
          variant="contained"
          color="primary"
          type="submit"
          className="petitionSubmit"
        >
          Submit
        </Button>
      </form>
      <Typography variant="h4">Current Petitions</Typography>
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Title</TableCell>
              <TableCell>Signatures</TableCell>
              <TableCell>Goal</TableCell>
              <TableCell>Signed By</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {petitions.map((petition) => (
              <TableRow key={petition.id}>
                <TableCell>{petition.title}</TableCell>
                <TableCell>{petition.signatures}</TableCell>
                <TableCell>{petition.goal}</TableCell>
                <TableCell>
                  <Button
                    variant="contained"
                    color="primary"
                    onClick={() => handleViewUsers(petition.users)}
                  >
                    View Users
                  </Button>
                </TableCell>
                <TableCell>
                  <Button
                    variant="contained"
                    color="error"
                    onClick={() => handleDelete(petition.id)}
                  >
                    <FontAwesomeIcon icon={faTrash} />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      <Dialog open={openModal} onClose={handleCloseModal}>
        <DialogTitle>Users</DialogTitle>
        <DialogContent>
          {selectedUsers.map((user) => (
            <Typography key={user.fullName}>{user.fullName}</Typography>
          ))}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseModal} color="primary">
            Close
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
}

export default AdminPetitions;
