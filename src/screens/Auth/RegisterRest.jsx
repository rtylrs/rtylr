import React from "react";
import PropTypes from "prop-types";
import { useNavigate } from "react-router-dom";
import notify from "../../lib/notify/notify";
import { useDispatch, useSelector } from "react-redux";
import {
  RegisterResaurantAsync,
  selectIsAuthenticated,
  selectLoading,
} from "../../redux/reducers/userSlice";

function RegisterUser() {
  const navigate = useNavigate();
  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const [name, setName] = React.useState("");
  const [refered, setRefered] = React.useState("");
  const dispatch = useDispatch();
  const isAuthenticated = useSelector(selectIsAuthenticated);
  const loading = useSelector(selectLoading);

  const handleRegister = () => {
    if (password === confirmPassword) {
      dispatch(RegisterResaurantAsync({ email, password, name, refered }));
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
    <div className="RegisterScreen-container">
      <div className="RegisterScreen">
        <h2 className="RegisterScreen-title">Sign Up</h2>
        <form onSubmit={handleRegister}>
          <input
            className="RegisterScreen-input"
            type="text"
            placeholder="Restaurant Name"
            id="Fname"
            required
            onChange={(e) => setName(e.target.value)}
          />
          <input
            className="RegisterScreen-input"
            type="email"
            placeholder="Email"
            id="email"
            required
            onChange={(e) => setEmail(e.target.value)}
          />
          <input
            className="RegisterScreen-input"
            placeholder="Password"
            type="password"
            id="pass"
            required
            onChange={(e) => setPassword(e.target.value)}
          />
          <input
            className="RegisterScreen-input"
            type="password"
            id="conPass"
            placeholder="Confirm Password"
            required
            onChange={(e) => setConfirmPassword(e.target.value)}
          />
          <input
            className="RegisterScreen-input"
            type="text"
            id="conPass"
            placeholder="Refered By (Optional)"
            required
            onChange={(e) => setRefered(e.target.value)}
          />
          <button
            className="RegisterScreen-button"
            onClick={() => {
              handleRegister();
            }}
            disabled={loading}
          >
            {loading ? "Loading..." : "Register"}
          </button>
        </form>

        <div className="LoginScreen-footer">
          <p>
            Already have an account?{" "}
            <span className="LoginScreen-link" onClick={() => {}}>
              Login Now.
            </span>
          </p>
          <p className="LoginScreen-link" onClick={() => {}}>
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
