import React from "react";
import { useNavigate } from "react-router-dom";
import notify from "../../lib/notify/notify";
import { useDispatch, useSelector } from "react-redux";
import {
  RegisterResaurantAsync,
  selectIsAuthenticated,
  selectLoading,
} from "../../redux/reducers/userSlice";

function RegisterRestaurant() {
  const navigate = useNavigate();
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
      dispatch(RegisterResaurantAsync({ email, password, name, address, phone, refered }));
    } else {
      notify.error("Password does not match");
    }
  };

  React.useEffect(() => {
    if (isAuthenticated) {
      navigate("/");
    }
  }, [isAuthenticated, navigate]);

  return (
    <div className="register-container">
      <div className="image-container">
        <img src="/images/Al-Abed_Mohammad2.jpg" alt="Restaurant" className="register-image" />
      </div>
      <div className="register-form">
        <h2>Sign Up</h2>
        <p>Create your account to get started</p>
        <form onSubmit={(e) => { e.preventDefault(); handleRegister(); }}>
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
          <button
            className="register-button"
            disabled={loading}
          >
            {loading ? "Loading..." : "Register"}
          </button>
        </form>
        <div className="register-footer">
          <p>
            Already have an account?{" "}
            <span className="login-link" onClick={() => navigate("/auth/login")}>
              Login Now.
            </span>
          </p>
          <p className="forgot-link" onClick={() => navigate("/auth/forgot-password")}>
            Forgot Password?
          </p>
        </div>
      </div>
    </div>
  );
}

export default RegisterRestaurant;
