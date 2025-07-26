import React, { useState } from "react";

const FeedbackSection = () => {
  const [rating, setRating] = useState(0);
  const [comments, setComments] = useState("");
  const [name, setName] = useState("");
  const [contact, setContact] = useState("");

  const handleRatingClick = (value) => {
    setRating(value);
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    // Handle form submission logic here
  };

  return (
    <div className="feedback-container">
      <h2>We appreciate your comments</h2>
      <p>Rate your experience</p>
      <div className="rating-stars">
        {[1, 2, 3, 4, 5].map((value) => (
          <span
            key={value}
            className={`star ${rating >= value ? "selected" : ""}`}
            onClick={() => handleRatingClick(value)}
          >
            ★
          </span>
        ))}
      </div>
      <form onSubmit={handleSubmit}>
        <textarea
          placeholder="Your Comments.."
          value={comments}
          onChange={(e) => setComments(e.target.value)}
        />
        <input
          type="text"
          placeholder="Full Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <input
          type="text"
          placeholder="Phone or Email"
          value={contact}
          onChange={(e) => setContact(e.target.value)}
        />
        <button type="submit">Submit</button>
      </form>
    </div>
  );
};

export default FeedbackSection;
