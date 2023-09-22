import React from "react";

import { auth, db } from "../firebase";
import { GoogleAuthProvider } from "firebase/auth";
import { Modal } from "@mui/material";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faClose } from "@fortawesome/free-solid-svg-icons";
import { faGoogle } from "@fortawesome/free-brands-svg-icons";

// Phone input
import PropTypes from "prop-types";
import ShowAlert from "../utils/showAlert";
import { motion } from "framer-motion";

function Register({ open, close }) {
  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const [name, setName] = React.useState("");
  const [refered, setRefered] = React.useState("");
  const provider = new GoogleAuthProvider();

  const handleRegister = (e) => {
    e.preventDefault();

    if (password === confirmPassword) {
      auth
        .createUserWithEmailAndPassword(email, password)
        .then((auth) => {
          auth.user.sendEmailVerification();

          db.collection("Users")
            .doc(auth.user.uid)
            .set(
              {
                email: email,
                name: name,
                isAdmin: false,
                Description: "",
                Location: "",
                About: "",
                image:
                  "https://firebasestorage.googleapis.com/v0/b/munasabti-32684.appspot.com/o/config%2Fprofile_default.jpg?alt=media&token=63f3178f-1b7a-4c4f-bd36-9f6da5893794",
                Socials: {
                  Instagram: "",
                  Website: "",
                  Facebook: "",
                  LinkedIn: "",
                  Twitter: "",
                },
                verified: false,
                tier: 0,
                refered: refered,
              },
              { merge: true }
            )
            .catch((e) => {
              alert(e.message);
            });

          auth.user
            .updateProfile({
              displayName: name,
              photoURL:
                "https://firebasestorage.googleapis.com/v0/b/munasabti-32684.appspot.com/o/config%2Fprofile_default.jpg?alt=media&token=63f3178f-1b7a-4c4f-bd36-9f6da5893794",
            })
            .then(() => {
              ShowAlert("Registered Successfully", "success");
              close();
            });
        })
        .catch((e) => {
          alert(e.message);
        });
    } else {
      ShowAlert("Passwords do not match", "error");
    }
  };

  const handleRegisterWithGoogle = () => {
    auth.signInWithPopup(provider).then((result) => {
      db.collection("Users")
        .doc(result.user.uid)
        .get()
        .then((doc) => {
          if (doc.exists) {
            close();
          } else {
            db.collection("Users")
              .doc(result.user.uid)
              .set({
                email: result.user.email,
                name: result.user.displayName,
                isAdmin: false,
                Description: "",
                Location: "",
                image:
                  "https://firebasestorage.googleapis.com/v0/b/munasabti-32684.appspot.com/o/config%2Fprofile_default.jpg?alt=media&token=63f3178f-1b7a-4c4f-bd36-9f6da5893794",
                Socials: {
                  Instagram: "",
                  Website: "",
                  Facebook: "",
                  LinkedIn: "",
                  Twitter: "",
                },
                verified: false,
                tier: 0,
                refered: refered,
              })
              .then(() => {
                auth.currentUser
                  .updateProfile({
                    displayName: result.user.displayName,
                    photoURL:
                      "https://firebasestorage.googleapis.com/v0/b/munasabti-32684.appspot.com/o/config%2Fprofile_default.jpg?alt=media&token=63f3178f-1b7a-4c4f-bd36-9f6da5893794",
                  })
                  .then(() => {
                    ShowAlert("Registered Successfully", "success");

                    db.collection("mail").add({
                      to: auth.currentUser.email,
                      message: {
                        subject: "Hello from Sanity!",
                        text: "Thanks for signing up!",
                        html: `<p>Thanks for signing up!</p>`,
                      },
                    });
                    close();
                  });
              })
              .catch((e) => {
                alert(e.message);
              });
          }
        })
        .catch((e) => {
          console.log(e.message);
        });
    });
  };

  return (
    <Modal open={open} onClose={close}>
      <motion.div
        className="Register-modal-container"
        initial={{ y: -1000 }}
        animate={{ y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <FontAwesomeIcon
          icon={faClose}
          className="LoginModal-close-icon"
          onClick={() => {
            close(true);
          }}
        />
        <h2 className="Register-title">Sign Up</h2>
        <form onSubmit={handleRegister}>
          <input
            className="RegisterModal-input"
            type="text"
            placeholder="Full Name"
            id="Fname"
            required
            onChange={(e) => setName(e.target.value)}
          />
          <input
            className="RegisterModal-input"
            type="email"
            placeholder="Email"
            id="email"
            required
            onChange={(e) => setEmail(e.target.value)}
          />
          <input
            className="RegisterModal-input"
            placeholder="Password"
            type="password"
            id="pass"
            required
            onChange={(e) => setPassword(e.target.value)}
          />
          <input
            className="RegisterModal-input"
            type="password"
            id="conPass"
            placeholder="Confirm Password"
            required
            onChange={(e) => setConfirmPassword(e.target.value)}
          />
          <input
            className="RegisterModal-input"
            type="text"
            id="conPass"
            placeholder="Refered By (Optional)"
            required
            onChange={(e) => setRefered(e.target.value)}
          />
          <button
            className="RegisterModal-button"
            onClick={() => {
              handleRegister();
            }}
          >
            {"Sign Up"}
          </button>
          <div className="LoginModal-social-media">
            <button
              className="LoginModal-google-button"
              onClick={() => {
                handleRegisterWithGoogle();
              }}
            >
              <FontAwesomeIcon icon={faGoogle} size="lg" />
              Sign in with Google
            </button>
          </div>
        </form>
      </motion.div>
    </Modal>
  );
}

Register.propTypes = {
  open: PropTypes.bool.isRequired,
  close: PropTypes.func.isRequired,
};

export default Register;
