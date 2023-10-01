import React, { useState } from "react";
import Modal from "@mui/material/Modal";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import PropTypes from "prop-types";
import { db, firebase } from "../firebase";

const PetitionModal = ({ isOpen, onClose, petitionId }) => {
  const [email, setEmail] = useState("");
  const [fullName, setFullName] = useState("");

  const handleSubmit = () => {
    db.collection("petitions")
      .doc(petitionId)
      .update({
        signatures: firebase.firestore.FieldValue.increment(1),
        users: firebase.firestore.FieldValue.arrayUnion({ email, fullName }),
      })
      .then(() => {
        setEmail("");
        setFullName("");

        onClose();
      });
  };

  return (
    <Modal
      open={isOpen}
      onClose={onClose}
      aria-labelledby="modal-title"
      aria-describedby="modal-description"
    >
      <Box className="modal-content">
        <h2 id="modal-title">Sign the Petition</h2>
        <form>
          <div className="form-group">
            <TextField
              label="Email"
              variant="outlined"
              type="email"
              fullWidth
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          <div className="form-group">
            <TextField
              label="Full Name"
              variant="outlined"
              fullWidth
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
              required
            />
          </div>
          <div className="button-group">
            <Button
              variant="contained"
              color="primary"
              type="submit"
              style={{
                marginRight: "1rem",
              }}
              onClick={() => {
                console.log("clicked");
                handleSubmit();
              }}
            >
              Submit
            </Button>
            <Button variant="outlined" color="error" onClick={onClose}>
              Close
            </Button>
          </div>
        </form>
      </Box>
    </Modal>
  );
};

PetitionModal.propTypes = {
  isOpen: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  petitionId: PropTypes.string.isRequired,
};

export default PetitionModal;
