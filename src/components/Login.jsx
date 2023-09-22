import React from "react";
import { auth, db } from "../firebase";
import { useNavigate } from "react-router-dom";
import { GoogleAuthProvider } from "firebase/auth";
import ShowAlert from "../utils/showAlert";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faClose } from "@fortawesome/free-solid-svg-icons";
import { faGoogle } from "@fortawesome/free-brands-svg-icons";
import { Modal } from "@mui/material";
import { motion } from "framer-motion";
import { useDispatch } from "react-redux";
import { login } from "../redux/reducers/userSlice";

import PropTypes from "prop-types";

function Login({ open, close, openRegister }) {
  const navigate = useNavigate();
  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");
  const dispatch = useDispatch();

  const provider = new GoogleAuthProvider();

  const handleSubmit = () => {
    auth
      .signInWithEmailAndPassword(email, password)
      .then((userCredential) => {
        const user = userCredential.user;
        if (user) {
          db.collection("Users")
            .doc(user.uid)
            .get()
            .then((doc) => {
              const user = Object.assign({}, doc.data(), {
                uid: doc.id,
              });
              dispatch(login(user));
              close(true);

              ShowAlert(`Welcome Back!`, "success");
            })
            .catch((e) => {
              alert(e.message);
            });
        }
      })
      .catch((e) => {
        alert(e.message);
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

          ShowAlert(`Welcome Back!`, "success");
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
      <motion.div
        className={`LoginModal-container ${open ? "show" : ""}`}
        onClick={() => {
          close(true);
        }}
        initial={{ y: -1000 }}
        animate={{ y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <div className="LoginModal" onClick={(e) => e.stopPropagation()}>
          <FontAwesomeIcon
            icon={faClose}
            className="LoginModal-close-icon"
            onClick={() => {
              close(true);
            }}
          />
          <h2 className="LoginModal-title">Welcome Back</h2>
          <p className="LoginModal-description">
            Sign in to your existing account to continue.
          </p>
          <input
            type="text"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="LoginModal-input"
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="LoginModal-input"
          />
          <button
            className="LoginModal-button"
            onClick={() => {
              handleSubmit();
            }}
          >
            Login
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
          <div className="LoginModal-footer">
            <p>
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
            </p>
            <p
              className="LoginModal-link"
              onClick={() => {
                navigate("/auth/password-reset");
              }}
            >
              Forgot Password?
            </p>
          </div>
        </div>
      </motion.div>
    </Modal>
  );
}

Login.propTypes = {
  open: PropTypes.bool.isRequired,
  close: PropTypes.func.isRequired,
  openRegister: PropTypes.func.isRequired,
};

export default Login;
