import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import notify from "../../lib/notify/notify";
import { useDispatch, useSelector } from "react-redux";
import { GoogleAuthProvider, signInWithPopup, getIdToken } from "firebase/auth";
import { auth } from "../../firebase";
import {
  getUserInfo,
  registerRestaurantAsync,
  selectIsAuthenticated,
  selectLoading,
  selectError,
  selectRole,
  restaurantWithGoogleAsync,
  clearError,
} from "../../redux/reducers/userSlice";
import { RootState, AppDispatch } from "../../redux/store";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faGoogle } from "@fortawesome/free-brands-svg-icons";
import VbitezButton from "../../components/VbitezButton/VbitezButton";

const Register: React.FC = () => {
  const navigate = useNavigate();
  const provider = new GoogleAuthProvider();
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [confirmPassword, setConfirmPassword] = useState<string>("");
  const [name, setName] = useState<string>("");
  const [referred, setReferred] = useState<string>("");
  const dispatch = useDispatch<AppDispatch>();
  const isAuthenticated = useSelector((state: RootState) =>
    selectIsAuthenticated(state)
  );
  const loading = useSelector((state: RootState) => selectLoading(state));
  const error = useSelector((state: RootState) => selectError(state));
  const role = useSelector((state: RootState) => selectRole(state));

  const handleRegister = () => {
    if (password === confirmPassword) {
      dispatch(
        registerRestaurantAsync({
          email,
          password,
          name,
          referred,
        })
      );
    } else {
      notify.error("Password does not match");
    }
  };

  const handleRegisterWithGoogle = async () => {
    try {
      const result = await signInWithPopup(auth, provider);
      const user = result.user;
      const idToken = await getIdToken(user);

      dispatch(restaurantWithGoogleAsync({ idToken }));
    } catch (error) {
      notify.error((error as any).message);
    }
  };

  useEffect(() => {
    // Clear any existing errors when component mounts
    dispatch(clearError());
  }, [dispatch]);

  useEffect(() => {
    if (isAuthenticated) {
      dispatch(getUserInfo());
      if (role === "business") {
        navigate("/auth/questionnaire");
      } else {
        navigate("/");
      }
    }
  }, [isAuthenticated, navigate]);

  return (
    <div className="auth-container">
      <div className="auth-form-container">
        <div className="auth-form">
          <h2>Sign Up</h2>
          <p>Create your free account to get started</p>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              handleRegister();
            }}
          >
            <label
              htmlFor="register-name"
              style={{
                fontWeight: 600,
                marginBottom: 4,
                alignSelf: "flex-start",
              }}
            >
              Restaurant Name
            </label>
            <input
              id="register-name"
              className="auth-input"
              type="text"
              placeholder="Restaurant Name"
              required
              onChange={(e) => setName(e.target.value)}
            />
            <label
              htmlFor="register-email"
              style={{
                fontWeight: 600,
                marginBottom: 4,
                alignSelf: "flex-start",
              }}
            >
              Email
            </label>
            <input
              id="register-email"
              className="auth-input"
              type="email"
              placeholder="Email"
              required
              onChange={(e) => setEmail(e.target.value)}
            />
            <label
              htmlFor="register-password"
              style={{
                fontWeight: 600,
                marginBottom: 4,
                alignSelf: "flex-start",
              }}
            >
              Password
            </label>
            <input
              id="register-password"
              className="auth-input"
              type="password"
              placeholder="Password"
              required
              onChange={(e) => setPassword(e.target.value)}
            />
            <label
              htmlFor="register-confirm-password"
              style={{
                fontWeight: 600,
                marginBottom: 4,
                alignSelf: "flex-start",
              }}
            >
              Confirm Password
            </label>
            <input
              id="register-confirm-password"
              className="auth-input"
              type="password"
              placeholder="Confirm Password"
              required
              onChange={(e) => setConfirmPassword(e.target.value)}
            />

            <div style={{ width: "100%", marginBottom: 12 }}>
              <small style={{ color: "#888", fontSize: 13 }}>
                By pressing Register, you agree to our{" "}
                <a
                  href="/terms"
                  target="_blank"
                  rel="noopener noreferrer"
                  style={{ textDecoration: "underline", color: "inherit" }}
                >
                  Terms and Services
                </a>
                .
              </small>
            </div>

            {error && (
              <p
                style={{
                  color: "#dc2626",
                  fontSize: 14,
                  marginBottom: 16,
                  textAlign: "center",
                }}
              >
                {typeof error === "string"
                  ? error
                  : (error as any)?.message || "An error occurred"}
              </p>
            )}

            <VbitezButton
              variant="primary"
              size="md"
              fullWidth
              loading={loading}
              type="submit"
            >
              Register
            </VbitezButton>

            <div className="auth-separator">
              <span>or</span>
            </div>

            <VbitezButton
              variant="outline"
              size="md"
              fullWidth
              type="button"
              onClick={handleRegisterWithGoogle}
              prefix={<FontAwesomeIcon icon={faGoogle} />}
            >
              Sign in with Google
            </VbitezButton>
          </form>

          <div className="auth-footer">
            <p>
              Already have an account?{" "}
              <span
                className="auth-link"
                onClick={() => navigate("/auth/login")}
              >
                Login Now.
              </span>
            </p>
            <p
              className="auth-link"
              onClick={() => navigate("/auth/forgot-password")}
            >
              Forgot Password?
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Register;
