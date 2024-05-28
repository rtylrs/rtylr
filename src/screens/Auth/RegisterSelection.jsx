import React from "react";
import { useNavigate } from "react-router-dom";

function RegisterSelection() {
  const navigate = useNavigate();

  return (
    <div className="register-selection-container">
      <h2>Register as</h2>
      <div className="register-options">
        <button onClick={() => navigate("/auth/register/user")} className="register-option-button">User</button>
        <button onClick={() => navigate("/auth/register/restaurant")} className="register-option-button">Restaurant</button>
      </div>
    </div>
  );
}

export default RegisterSelection;
