import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import {
  loginAsync,
  selectIsAuthenticated,
  selectLoading,
  selectError,
  selectRole,
} from "../../redux/reducers/userSlice";
import VbitezButton from "../../components/Buttons/Buttons";

function Login() {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const isAuthenticated = useSelector(selectIsAuthenticated);
  const loading = useSelector(selectLoading);
  const error = useSelector(selectError);
  const role = useSelector(selectRole);

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const handleSubmit = () => {
    console.log(email, password);
    dispatch(loginAsync({ email, password }));
  };

  React.useEffect(() => {
    if (isAuthenticated) {
      navigate("/");
    }
  }, [isAuthenticated, navigate]);

  return (
    <div className="login-container">
      <div className="login-form">
        <h3 className="login-form-title">Email address</h3>
        <input
          type="email"
          placeholder="Email address"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="login-input"
        />
        <h3 className="login-form-title">Password</h3>
        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="login-input"
        />
        {error && <p className="login-error">{error}</p>}
        <button
          className="login-button"
          onClick={handleSubmit}
          disabled={loading}
        >
          {loading ? "Loading..." : "Log in"}
        </button>
        <p className="login-forgot-password">Forgot your password?</p>
        <div className="login-separator">
          <span>or</span>
        </div>
        <button className="login-social-button">
          <img src="/images/google-icon.png" alt="Google Icon" />
          Log in with Google
        </button>
      
        <p className="login-register">
          Don’t have an account yet?{" "}
          <span
            className="login-link"
            onClick={() => navigate("/auth/register")}
          >
            Register
          </span>
        </p>
      </div>
    </div>
  );
}

export default Login;
