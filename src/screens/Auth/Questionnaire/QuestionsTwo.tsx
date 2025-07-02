import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import VbitezButton from "../../../components/VbitezButton/VbitezButton";

const QuestionsTwo: React.FC = () => {
  const [selectedOption, setSelectedOption] = useState<string>("");
  const navigate = useNavigate();

  const handleOptionChange = (option: string) => {
    setSelectedOption(option);
  };

  const handleContinue = () => {
    navigate("/app/business/");
  };

  const handleSkip = () => {
    navigate("/app/business/");
  };

  const handleBack = () => {
    navigate("/auth/questionnaire/info/question-1");
  };

  return (
    <div className="questionaire-container">
      <form className="questionaire-form">
        <h1>What best describes you?</h1>
        <p>This will enable us to customize the system to meet your needs.</p>

        <div className="options-group">
          <label className="option">
            <input
              type="radio"
              value="Restaurant owner"
              checked={selectedOption === "Restaurant owner"}
              onChange={() => handleOptionChange("Restaurant owner")}
            />
            Restaurant owner
          </label>
          <label className="option">
            <input
              type="radio"
              value="Restaurant manager"
              checked={selectedOption === "Restaurant manager"}
              onChange={() => handleOptionChange("Restaurant manager")}
            />
            Restaurant manager
          </label>
          <label className="option">
            <input
              type="radio"
              value="Marketing specialist"
              checked={selectedOption === "Marketing specialist"}
              onChange={() => handleOptionChange("Marketing specialist")}
            />
            Marketing specialist
          </label>
          <label className="option">
            <input
              type="radio"
              value="Website developer"
              checked={selectedOption === "Website developer"}
              onChange={() => handleOptionChange("Website developer")}
            />
            Website developer
          </label>
          <label className="option">
            <input
              type="radio"
              value="I'm looking to place a food order"
              checked={selectedOption === "I'm looking to place a food order"}
              onChange={() =>
                handleOptionChange("I'm looking to place a food order")
              }
            />
            I’m looking to place a food order
          </label>
          <label className="option">
            <input
              type="radio"
              value="Other"
              checked={selectedOption === "Other"}
              onChange={() => handleOptionChange("Other")}
            />
            Other
          </label>
        </div>
        <div className="form-actions">
          <VbitezButton
            onClick={handleContinue}
            styles={{
              width: "100%",
            }}
          >
            Continue
          </VbitezButton>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              marginTop: "10px",
            }}
          >
            <span onClick={handleBack} className="back-button">
              ← Back
            </span>
            <span onClick={handleSkip} className="skip-button">
              Skip
            </span>
          </div>
        </div>
      </form>
      <div className="questionaire-image">
        <img src="/svg/info.svg" alt="Logo" className="logo" />
      </div>
    </div>
  );
};

export default QuestionsTwo;
