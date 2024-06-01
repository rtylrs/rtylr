import React from "react";
import { useNavigate } from "react-router-dom";
import { auth, db } from "../../firebase";
import { GoogleAuthProvider } from "firebase/auth";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faGoogle } from "@fortawesome/free-brands-svg-icons";
import PropTypes from "prop-types";

function RegisterUser() {
  const navigate = useNavigate();
  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const [name, setName] = React.useState("");
  const provider = new GoogleAuthProvider();

  const handleRegister = () => {
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
                isVendor: false,
                isContributor: false,
                Description: "",
                Location: "",
                About: "",
                image:
                  "https://firebasestorage.googleapis.com/v0/b/mkunasabti-32684.appspot.com/o/config%2Fprofile_default.jpg?alt=media&token=63f3178f-1b7a-4c4f-bd36-9f6da5893794",
                Socials: {
                  Instagram: "",
                  Website: "",
                  Facebook: "",
                  LinkedIn: "",
                  Twitter: "",
                },
                verified: false,
                tier: 0,
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
              navigate("/");
            });
        })
        .catch((e) => {
          alert(e.message);
        });
    } else {
      alert("Passwords do not match");
    }
  };

  const handleRegisterWithGoogle = () => {
    auth.signInWithPopup(provider).then((result) => {
      db.collection("Users")
        .doc(result.user.uid)
        .get()
        .then((doc) => {
          if (doc.exists) {
            navigate("/");
          } else {
            db.collection("Users")
              .doc(result.user.uid)
              .set({
                email: result.user.email,
                name: result.user.displayName,
                isAdmin: false,
                isVendor: false,
                isContributor: false,
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
              })
              .then(() => {
                auth.currentUser
                  .updateProfile({
                    displayName: result.user.displayName,
                    photoURL:
                      "https://firebasestorage.googleapis.com/v0/b/munasabti-32684.appspot.com/o/config%2Fprofile_default.jpg?alt=media&token=63f3178f-1b7a-4c4f-bd36-9f6da5893794",
                  })
                  .then(() => {
                    db.collection("mail").add({
                      to: auth.currentUser.email,
                      message: {
                        subject: "Hello from Voxify!",
                        text: "Thanks for signing up!",
                        html: `<p>Thanks for signing up!</p>`,
                      },
                    });
                    navigate("/");
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
    <div className="register-user-container">
      <div className="register-user-image-container">
        <img src="/images/Al-Abed_Mohammad2.jpg" alt="Register" className="register-user-image" />
      </div>
      <div className="register-user-form">
        <img src="/public/logo.png" alt="Logo" className="register-user-logo" />
        <h2>Sign Up</h2>
        <p>Create your user account to get started</p>
        <form onSubmit={(e) => { e.preventDefault(); handleRegister(); }}>
          <input
            className="register-user-input"
            type="text"
            placeholder="Full Name"
            required
            onChange={(e) => setName(e.target.value)}
          />
          <input
            className="register-user-input"
            type="email"
            placeholder="Email"
            required
            onChange={(e) => setEmail(e.target.value)}
          />
          <input
            className="register-user-input"
            type="password"
            placeholder="Password"
            required
            onChange={(e) => setPassword(e.target.value)}
          />
          <input
            className="register-user-input"
            type="password"
            placeholder="Confirm Password"
            required
            onChange={(e) => setConfirmPassword(e.target.value)}
          />
          <button className="register-user-button" disabled={false}>
            Sign Up
          </button>
          <div className="register-user-social-media">
            <button className="register-user-google-button" onClick={handleRegisterWithGoogle}>
              <FontAwesomeIcon icon={faGoogle} size="lg" />
              Sign in with Google
            </button>
          </div>
        </form>
        <div className="register-user-footer">
          <p>
            Already have an account?{" "}
            <span className="register-user-login-link" onClick={() => navigate("/auth/login")}>
              Login Now.
            </span>
          </p>
          <p className="register-user-forgot-link" onClick={() => navigate("/auth/forgot-password")}>
            Forgot Password?
          </p>
        </div>
      </div>
    </div>
  );
}

RegisterUser.propTypes = {
  open: PropTypes.bool.isRequired,
  close: PropTypes.func.isRequired,
};

export default RegisterUser;
