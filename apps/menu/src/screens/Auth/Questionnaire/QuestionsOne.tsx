import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from '@voxire/components';

const QuestionsOne: React.FC = () => {
  const [website, setWebsite] = useState<string>("");
  const [thirdPartyLink, setThirdPartyLink] = useState<string>("");
  const navigate = useNavigate();

  const handleContinue = () => {
    navigate("/auth/questionnaire/info/question-2");
  };

  const handleSkip = () => {
    navigate("/auth/questionnaire/info/question-2");
  };

  const handleBack = () => {
    // Implement the logic to handle the back action
    navigate(-1); // Go back to the previous page
  };

  return (
    <div className="questionaire-container">
      <form className="questionaire-form">
        <h1>Do you have a restaurant website?</h1>
        <div className="form-group">
          <input
            type="text"
            id="website"
            placeholder="Enter your current website address"
            value={website}
            onChange={(e) => setWebsite(e.target.value)}
          />
        </div>
        <h1>Do you sell online on a third party app?</h1>
        <div
          className="form-group"
          style={{ display: "flex", flexDirection: "column", gap: "5px" }}
        >
          <input
            type="text"
            id="thirdPartyLink"
            placeholder="Enter a link to your ordering page"
            value={thirdPartyLink}
            onChange={(e) => setThirdPartyLink(e.target.value)}
          />
          <small>Example: ubereats.com/en/store/your-restaurant</small>
        </div>
        <div className="form-actions">
          <Button
            onClick={handleContinue}
            styles={{
              width: "100%",
            }}
          >
            Continue
          </Button>
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

export default QuestionsOne;
