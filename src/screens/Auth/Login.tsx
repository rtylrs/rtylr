import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import {
  loginAsync,
  selectIsAuthenticated,
  selectLoading,
  selectError,
  selectRole,
} from "../../redux/reducers/userSlice";
import { RootState, AppDispatch } from "../../redux/store";

const Login: React.FC = () => {
  const navigate = useNavigate();
  const dispatch: AppDispatch = useDispatch();
  const isAuthenticated = useSelector((state: RootState) =>
    selectIsAuthenticated(state)
  );
  const loading = useSelector((state: RootState) => selectLoading(state));
  const error = useSelector((state: RootState) => selectError(state));
  const role = useSelector((state: RootState) => selectRole(state));

  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");

  const handleSubmit = () => {
    dispatch(loginAsync({ email, password }));
  };

  useEffect(() => {
    if (isAuthenticated) {
      if (role === "business") {
        navigate("/app/business");
      } else {
        navigate("/");
      }
    }
  }, [isAuthenticated, navigate, role]);

  return (
    <div className="login-container">
      <div className="login-form">
        <img src="/logo.png" alt="Logo" className="logo" />
        <h2>Welcome back!!</h2>
        <p>Please enter your details.</p>
        <input
          type="email"
          placeholder="Enter your email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="login-input"
        />
        <input
          type="password"
          placeholder="Enter your password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="login-input"
        />
        {error && <p className="login-error">{error}</p>}
        <div className="login-options">
          <span className="forgot-password">Forgot password?</span>
        </div>
        <div className="login-button-wrapper">
          <button className="login-button" onClick={handleSubmit}>
            {loading ? "Loading..." : "Login"}
          </button>
        </div>
        <div className="login-separator">
          <span>or</span>
        </div>
        <button className="login-social-button">
          <img src="/images/google-icon.png" alt="Google Icon" />
          Log in with Google
        </button>
        <p className="signup-link">
          Don’t have an account?{" "}
          <span onClick={() => navigate("/auth/register")}>Sign up</span>
        </p>
      </div>
      <div className="login-image">
        <img src="/images/LoginBG.jpg" alt="Logo" className="logo" />
      </div>
    </div>
  );
};

export default Login;
