import React, { useState } from "react";
import Modal from "@mui/material/Modal";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import PropTypes from "prop-types";
import { db } from "../firebase";

const PetitionModal = ({ isOpen, onClose, petitionId }) => {
  const [email, setEmail] = useState("");
  const [fullName, setFullName] = useState("");
  const [petition, setPetition] = useState([]);


  

  const handleSubmit = (e) => {
    e.preventDefault();
    db.collection("petitions")
      .doc(petitionId)
      .get()
      .then((doc) => {
        setPetition(doc.data());
      }
    );
    
    db.collection("petitions")
      .doc(petitionId)
      .update({
        signatures: petition.signatures + 1,
        users: [...petition.users, { email, fullName }]
      });

    console.log("im here");
    

    // db.collection("petitions")
    //   .doc(petitionId)
    //   .update({
    //     signatures: db.FieldValue.increment(1),
    //     users: db.FieldValue.arrayUnion({ email, fullName }),
    //   });

    onClose();
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
        <form onSubmit={handleSubmit}>
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
            <Button variant="contained" color="primary" type="submit">
              Submit
            </Button>
            <Button variant="outlined" color="secondary" onClick={onClose}>
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
};

export default PetitionModal;
