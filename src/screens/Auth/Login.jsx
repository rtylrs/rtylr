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
    //TODO: fix this

    if (isAuthenticated) {
      if (role === "admin") {
        // navigate("/manage");
        console.log("navigate to manage");
      }
      // navigate("/");
    }
  }, [isAuthenticated, navigate]);

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
