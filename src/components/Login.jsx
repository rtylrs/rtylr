import React from "react";
import { auth, db } from "../firebase";
import { useNavigate } from "react-router-dom";
import { GoogleAuthProvider } from "firebase/auth";
import ShowAlert from "../utils/showAlert";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faClose } from "@fortawesome/free-solid-svg-icons";
import { faGoogle } from "@fortawesome/free-brands-svg-icons";
import { Modal, Typography, TextField, Button } from "@mui/material";
import PropTypes from "prop-types";

function Login({ open, close, openRegister }) {
  const navigate = useNavigate();
  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");

  const provider = new GoogleAuthProvider();

  const handleSubmit = () => {
    db.collection("Users")
      .where("username", "==", email)
      .get()
      .then((querySnapshot) => {
        if (querySnapshot.docs.length > 0) {
          auth
            .signInWithEmailAndPassword(
              querySnapshot.docs[0].data().email,
              password
            )
            .then((user) => {
              db.collection("Users")
                .doc(user.user.uid)
                .get()
                .then(() => {
                  close(true);
                  ShowAlert("Logged In Successfully", "success");
                });
            })
            .catch((e) => {
              alert(e.message);
            });
        } else {
          auth
            .signInWithEmailAndPassword(email, password)
            .then((user) => {
              db.collection("Users")
                .doc(user.user.uid)
                .get()
                .then(() => {
                  close(true);
                  ShowAlert("Logged In Successfully", "success");
                });
            })
            .catch((e) => {
              alert(e.message);
            });
        }
      });
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
              .set(
                {
                  email: result.user.email,
                  name: result.user.displayName,
                  isAdmin: false,
                  isVendor: false,
                  Description: "",
                  Location: "",
                  Phone: "",
                  image:
                    "https://firebasestorage.googleapis.com/v0/b/munasabti-32684.appspot.com/o/config%2Fprofile_default.jpg?alt=media&token=63f3178f-1b7a-4c4f-bd36-9f6da5893794",
                  Socials: {
                    Instagram: "",
                    Website: "",
                    Facebook: "",
                    LinkedIn: "",
                    Twitter: "",
                  },
                  VerifiedVendor: false,
                },
                { merge: true }
              )
              .then(() => {
                auth.currentUser.updateProfile({
                  displayName: result.user.displayName,
                  photoURL:
                    "https://firebasestorage.googleapis.com/v0/b/munasabti-32684.appspot.com/o/config%2Fprofile_default.jpg?alt=media&token=63f3178f-1b7a-4c4f-bd36-9f6da5893794",
                });
              });
          }

          ShowAlert("Logged In Successfully", "success");
          close();
        })
        .catch((e) => {
          console.log(e.message);
        });
    });
  };

  return (
    <Modal
      open={open}
      onClose={() => {
        close(true);
      }}
    >
      <div className="LoginModal-container">
        <FontAwesomeIcon
          icon={faClose}
          className="LoginModal-close-icon"
          onClick={() => {
            close(true);
          }}
          size={20}
        />
        <Typography variant="h5" className="LoginModal-title">
          Welcome Back
        </Typography>
        <Typography variant="body1" className="LoginModal-description">
          Sign in to your existing account to continue.
        </Typography>
        <TextField
          label="Username or Email"
          variant="outlined"
          fullWidth
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          style={{
            marginTop: "10px",
          }}
        />
        <TextField
          label="Password"
          variant="outlined"
          type="password"
          fullWidth
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          style={{
            marginTop: "10px",
          }}
        />
        <Button
          style={{
            marginTop: "20px",
          }}
          variant="contained"
          color="primary"
          onClick={() => {
            handleSubmit();
          }}
          fullWidth
        >
          Login
        </Button>
        <div className="LoginModal-social-media">
          <Button
            variant="contained"
            color="primary"
            onClick={() => {
              handleRegisterWithGoogle();
            }}
            fullWidth
          >
            <FontAwesomeIcon icon={faGoogle} size="lg" />
            Sign in with Google
          </Button>
        </div>
        <div className="LoginModal-footer">
          <Typography variant="body2">
            Don’t have an account?{" "}
            <span
              className="LoginModal-link"
              onClick={() => {
                openRegister();
                close(true);
              }}
            >
              Sign Up Now.
            </span>
          </Typography>
          <Typography
            variant="body2"
            className="LoginModal-link"
            onClick={() => {
              navigate("/auth/password-reset");
            }}
          >
            Forgot Password?
          </Typography>
        </div>
      </div>
    </Modal>
  );
}

Login.propTypes = {
  open: PropTypes.bool.isRequired,
  close: PropTypes.func.isRequired,
  openRegister: PropTypes.func.isRequired,
};

export default Login;
