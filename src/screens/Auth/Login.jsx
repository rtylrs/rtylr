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
    dispatch(loginAsync({ email, password }));
  };

  useEffect(() => {
    if (isAuthenticated) {
      if (role === "admin") {
        navigate("/manage");
      } else if (role === "business") {
        navigate("/app/restaurant");
      } else {
        navigate("/");
      }
    }
  }, [isAuthenticated, navigate, role]);

  useEffect(() => {
    // Dynamically load the Lottie player script
    const script = document.createElement("script");
    script.src =
      "https://unpkg.com/@dotlottie/player-component@latest/dist/dotlottie-player.mjs";
    script.type = "module";
    document.body.appendChild(script);
  }, []);

  return (
    <div className="login-container">
      <div className="lottie-container">
        <dotlottie-player
          src="https://lottie.host/13436e76-0a4b-464c-9e28-bb720ca012cf/JDvAdmkONp.json"
          background="transparent"
          speed="1"
          style={{ width: "700px", height: "700px" }}
          loop
          autoplay
        ></dotlottie-player>
      </div>
      <div className="login-form">
        <img src="/public/logo.png" alt="Logo" className="logo" />
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
          <button
            className="login-button"
            onClick={handleSubmit}
            disabled={loading}
          >
            {loading ? "Loading..." : "Sign In"}
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
    </div>
  );
}

export default Login;
