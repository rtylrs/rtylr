import React, { useState } from "react";
import { auth, db } from "../../firebase";
import { useNavigate } from "react-router-dom";
import { GoogleAuthProvider } from "firebase/auth";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faGoogle } from "@fortawesome/free-brands-svg-icons";
import { useDispatch } from "react-redux";
import { login } from "../../redux/reducers/userSlice";

function Login() {
  const navigate = useNavigate();
  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");
  const dispatch = useDispatch();
  const [error, setError] = useState(null);

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
              navigate("/");
            })
            .catch((e) => {
              setError(e.message);
            });
        }
      })
      .catch((e) => {
        setError(e.message);
      });
  };

  const handleRegisterWithGoogle = () => {
    auth.signInWithPopup(provider).then((result) => {
      db.collection("Users")
        .doc(result.user.uid)
        .get()
        .then((doc) => {
          if (doc.exists) {
            const user = Object.assign({}, doc.data(), {
              uid: doc.id,
            });
            dispatch(login(user));
            navigate("/");
          } else {
            db.collection("Users")
              .doc(result.user.uid)
              .set(
                {
                  email: result.user.email,
                  name: result.user.displayName,
                  isAdmin: false,
                  isVendor: false,
                  isContributor: false,
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
                auth.currentUser
                  .updateProfile({
                    displayName: result.user.displayName,
                    photoURL:
                      "https://firebasestorage.googleapis.com/v0/b/munasabti-32684.appspot.com/o/config%2Fprofile_default.jpg?alt=media&token=63f3178f-1b7a-4c4f-bd36-9f6da5893794",
                  })
                  .then(() => {
                    const user = Object.assign({}, doc.data(), {
                      uid: doc.id,
                    });
                    dispatch(login(user));
                    navigate("/");
                  });
              });
          }

        })
        .catch((e) => {
          setError(e.message);
        });
    });
  };

  return (
    <div className="LoginScreen-container">
      <div className="LoginScreen">
        <h2 className="LoginScreen-title">Welcome Back</h2>
        <p className="LoginScreen-description">
          Sign in to your existing account to continue.
        </p>
        <input
          type="text"
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="LoginScreen-input"
        />
        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="LoginScreen-input"
        />
        {error && <p className="LoginScreen-error">{error}</p>}
        <button
          className="LoginScreen-button"
          onClick={() => {
            handleSubmit();
          }}
        >
          Login
        </button>
        <div className="LoginScreen-social-media">
          <button
            className="LoginScreen-google-button"
            onClick={() => {
              handleRegisterWithGoogle();
            }}
          >
            <FontAwesomeIcon icon={faGoogle} size="lg" />
            Sign in with Google
          </button>
        </div>
        <div className="LoginScreen-footer">
          <p>
            Don’t have an account?{" "}
            <span
              className="LoginScreen-link"
              onClick={() => {
                navigate("/auth/register");
              }}
            >
              Sign Up Now.
            </span>
          </p>
          <p
            className="LoginScreen-link"
            onClick={() => {
              navigate("/auth/password-reset");
            }}
          >
            Forgot Password?
          </p>
        </div>
      </div>
    </div>
  );
}

export default Login;
