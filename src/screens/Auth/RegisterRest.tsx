import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import notify from "../../lib/notify/notify";
import { useDispatch, useSelector } from "react-redux";
import { GoogleAuthProvider, signInWithPopup } from "firebase/auth";
import { auth } from "../../firebase";
import {
  RegisterResaurantAsync,
  selectIsAuthenticated,
  selectLoading,
} from "../../redux/reducers/userSlice";
import { RootState, AppDispatch } from "../../redux/store";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faGoogle } from "@fortawesome/free-brands-svg-icons";

const RegisterRestaurant: React.FC = () => {
  const navigate = useNavigate();
  const provider = new GoogleAuthProvider();
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [confirmPassword, setConfirmPassword] = useState<string>("");
  const [name, setName] = useState<string>("");
  const [refered, setRefered] = useState<string>("");
  const dispatch = useDispatch<AppDispatch>();
  const isAuthenticated = useSelector((state: RootState) =>
    selectIsAuthenticated(state)
  );
  const loading = useSelector((state: RootState) => selectLoading(state));

  const handleRegister = () => {
    if (password === confirmPassword) {
      dispatch(
        RegisterResaurantAsync({
          email,
          password,
          name,
          refered,
        })
      );
    } else {
      notify.error("Password does not match");
    }
  };

  const handleRegisterWithGoogle = () => {
    signInWithPopup(auth, provider)
      .then((result) => {
        console.log(result);
      })
      .catch((error) => {
        notify.error(error.message);
      });
  };

  useEffect(() => {
    if (isAuthenticated) {
      navigate("/");
    }
  }, [isAuthenticated, navigate]);

  return (
    <div className="register-container">
      <div className="image-container">
        <video
          autoPlay
          muted
          loop
          src="/videos/Register.mp4"
          className="register-image"
        ></video>
      </div>
      <div className="register-form">
        <h2>Sign Up</h2>
        <p>Create your free account to get started</p>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleRegister();
          }}
        >
          <input
            className="register-input"
            type="text"
            placeholder="Restaurant Name"
            required
            onChange={(e) => setName(e.target.value)}
          />
          <input
            className="register-input"
            type="email"
            placeholder="Email"
            required
            onChange={(e) => setEmail(e.target.value)}
          />
          <input
            className="register-input"
            type="password"
            placeholder="Password"
            required
            onChange={(e) => setPassword(e.target.value)}
          />
          <input
            className="register-input"
            type="password"
            placeholder="Confirm Password"
            required
            onChange={(e) => setConfirmPassword(e.target.value)}
          />
          <input
            className="register-input"
            type="text"
            placeholder="Referred By (Optional)"
            onChange={(e) => setRefered(e.target.value)}
          />
          <div>
            <button className="register-button" disabled={loading}>
              {loading ? "Loading..." : "Register"}
            </button>
            <div className="register-user-social-media">
              <button
                type="button"
                className="register-user-google-button"
                onClick={handleRegisterWithGoogle}
              >
                <FontAwesomeIcon icon={faGoogle} size="lg" />
                Sign in with Google
              </button>
            </div>
          </div>
        </form>
        <div className="register-footer">
          <p>
            Already have an account?{" "}
            <span
              className="login-link"
              onClick={() => navigate("/auth/login")}
            >
              Login Now.
            </span>
          </p>
          <p
            className="forgot-link"
            onClick={() => navigate("/auth/forgot-password")}
          >
            Forgot Password?
          </p>
        </div>
      </div>
    </div>
  );
};

export default RegisterRestaurant;
