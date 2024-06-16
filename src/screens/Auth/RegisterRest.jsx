import React from "react";
import { useNavigate } from "react-router-dom";
import notify from "../../lib/notify/notify";
import { useDispatch, useSelector } from "react-redux";
import { GoogleAuthProvider } from "firebase/auth";
import { auth } from "../../firebase";
import {
  RegisterResaurantAsync,
  selectIsAuthenticated,
  selectLoading,
} from "../../redux/reducers/userSlice";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faGoogle } from "@fortawesome/free-brands-svg-icons";

function RegisterRestaurant() {
  const navigate = useNavigate();
  const provider = new GoogleAuthProvider();
  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const [name, setName] = React.useState("");
  const [address, setAddress] = React.useState("");
  const [phone, setPhone] = React.useState("");
  const [refered, setRefered] = React.useState("");
  const dispatch = useDispatch();
  const isAuthenticated = useSelector(selectIsAuthenticated);
  const loading = useSelector(selectLoading);

  const handleRegister = () => {
    if (password === confirmPassword) {
      dispatch(
        RegisterResaurantAsync({
          email,
          password,
          name,
          address,
          phone,
          refered,
        })
      );
    } else {
      notify.error("Password does not match");
    }
  };

  const handleRegisterWithGoogle = () => {
    auth
      .signInWithPopup(provider)
      .then((result) => {
        console.log(result);
      })
      .catch((error) => {
        notify.error(error.message);
      });
  };

  React.useEffect(() => {
    if (isAuthenticated) {
      navigate("/");
    }
  }, [isAuthenticated, navigate]);

  return (
    <div className="register-container">
      <div className="image-container">
        <img
          src="/images/Register.jpg"
          alt="Restaurant"
          className="register-image"
        />
      </div>
      <div className="register-form">
        <h2>Sign Up</h2>
        <p>Create your account to get started</p>
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
            type="text"
            placeholder="Address"
            required
            onChange={(e) => setAddress(e.target.value)}
          />
          <input
            className="register-input"
            type="text"
            placeholder="Phone Number"
            required
            onChange={(e) => setPhone(e.target.value)}
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
}

export default RegisterRestaurant;
