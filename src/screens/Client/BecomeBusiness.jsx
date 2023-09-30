import { useState } from "react";
import { Categories } from "../../data/categories"; // Import the categories data

function BecomeBusiness() {
  // Define state variables to store form data
  const [businessName, setBusinessName] = useState("");
  const [businessNameArabic, setbusinessNameArabic] = useState("");
  // const [selectedCategory, setSelectedCategory] = useState(""); // New state for selected category
  const [note, setNote] = useState("");

  // Handle form submission
  const handleSubmit = (e) => {
    e.preventDefault();

    // Perform actions like submitting data to a server, payment processing, etc.

    // Reset the form fields
    setBusinessName("");
    // setSelectedCategory("");
    setNote("");
  };

  return (
    <div className="BecomeBusiness-container">
      <h2>Become a Business</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="businessName">English Name:</label>
          <input
            type="text"
            id="businessName"
            value={businessName}
            onChange={(e) => setBusinessName(e.target.value)}
            required
          />
        </div>
        <div className="form-group">
          <label htmlFor="businessName">Arabic Name:</label>
          <input
            type="text"
            id="businessName"
            value={businessNameArabic}
            onChange={(e) => setbusinessNameArabic(e.target.value)}
            required
          />
        </div>
        <div className="form-group">
          <label htmlFor="businessName">Additional Note:</label>
          <textarea
            type="text"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            required
            rows={6}
          />
        </div>
        <div className="form-group">
          <label htmlFor="category">Country:</label>
          <select id="category" required>
            <option value="">Select a Country</option>
            <option value="Bahrain">Bahrain</option>
            <option value="Kuwait">Kuwait</option>
            <option value="Oman">Oman</option>
            <option value="Qatar">Qatar</option>
            <option value="Saudi Arabia">Saudi Arabia</option>
            <option value="United Arab Emirates">United Arab Emirates</option>
          </select>
        </div>
        <button type="submit">Submit Application</button>
      </form>
    </div>
  );
}

export default BecomeBusiness;
