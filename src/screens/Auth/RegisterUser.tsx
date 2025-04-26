import React from "react";
import { useNavigate } from "react-router-dom";
import { auth } from "../../firebase";
import { GoogleAuthProvider } from "firebase/auth";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faGoogle } from "@fortawesome/free-brands-svg-icons";

const RegisterUser: React.FC<{}> = () => {
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
        })
        .catch((e) => {
          alert(e.message);
        });
    } else {
      alert("Passwords do not match");
    }
  };

  const handleRegisterWithGoogle = () => {
    auth.signInWithPopup(provider).then((result) => {});
  };

  return (
    <div className="register-user-container">
      <div className="register-user-image-container">
        <img
          src="/images/Al-Abed_Mohammad2.jpg"
          alt="Register"
          className="register-user-image"
        />
      </div>
      <div className="register-user-form">
        <img src="/logo.png" alt="Logo" className="register-user-logo" />
        <h2>Sign Up</h2>
        <p>Create your user account to get started</p>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleRegister();
          }}
        >
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
            <button
              className="register-user-google-button"
              onClick={handleRegisterWithGoogle}
            >
              <FontAwesomeIcon icon={faGoogle} size="lg" />
              Sign in with Google
            </button>
          </div>
        </form>
        <div className="register-user-footer">
          <p>
            Already have an account?{" "}
            <span
              className="register-user-login-link"
              onClick={() => navigate("/auth/login")}
            >
              Login Now.
            </span>
          </p>
          <p
            className="register-user-forgot-link"
            onClick={() => navigate("/auth/forgot-password")}
          >
            Forgot Password?
          </p>
        </div>
      </div>
    </div>
  );
};

export default RegisterUser;
