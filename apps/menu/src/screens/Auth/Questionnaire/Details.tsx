import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useDispatch } from "react-redux";
import { AppDispatch } from "../../../redux/store";
import { Button } from '@voxire/shared';

const Details: React.FC = () => {
  const navigate = useNavigate();
  const dispatch: AppDispatch = useDispatch();

  const [restaurantName, setRestaurantName] = useState<string>("");
  const [firstName, setFirstName] = useState<string>("");
  const [lastName, setLastName] = useState<string>("");
  const [phoneNumber, setPhoneNumber] = useState<string>("");

  const handleSubmit = () => {
    navigate("/auth/questionnaire/info/question-1");
  };

  return (
    <div className="questionaire-container">
      <div className="questionaire-form">
        <h1>Enter your restaurant details</h1>
        <form>
          <div className="form-group">
            <input
              type="text"
              placeholder="Your restaurant name *"
              value={restaurantName}
              onChange={(e) => setRestaurantName(e.target.value)}
              required
            />
          </div>
          <div className="form-group">
            <input
              type="text"
              placeholder="Your first name *"
              value={firstName}
              onChange={(e) => setFirstName(e.target.value)}
              required
            />
            <input
              type="text"
              placeholder="Your last name *"
              value={lastName}
              onChange={(e) => setLastName(e.target.value)}
              required
            />
          </div>
          <div className="form-group">
            <input
              type="text"
              placeholder="Phone number *"
              value={phoneNumber}
              onChange={(e) => setPhoneNumber(e.target.value)}
              required
            />
          </div>
          <Button
            styles={{
              width: "100%",
            }}
            onClick={handleSubmit}
          >
            Continue
          </Button>
        </form>
        <p className="required-field">* Required field</p>
      </div>
      <div className="questionaire-image">
        <img src="/svg/info.svg" alt="Logo" className="logo" />
      </div>
    </div>
  );
};

export default Details;
