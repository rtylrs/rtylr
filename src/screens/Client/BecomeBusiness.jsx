import { useState } from "react";
import { Categories } from "../../data/categories"; // Import the categories data

function BecomeBusiness() {
  // Define state variables to store form data
  const [businessName, setBusinessName] = useState("");
  const [selectedCategory, setSelectedCategory] = useState(""); // New state for selected category

  // Handle form submission
  const handleSubmit = (e) => {
    e.preventDefault();

    // Perform actions like submitting data to a server, payment processing, etc.

    // Reset the form fields
    setBusinessName("");
    setSelectedCategory("");
  };

  return (
    <div className="BecomeBusiness-container">
      <h2>Become a Business</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="businessName">Business Name:</label>
          <input
            type="text"
            id="businessName"
            value={businessName}
            onChange={(e) => setBusinessName(e.target.value)}
            required
          />
        </div>
        <div className="form-group">
          <label htmlFor="category">Category:</label>
          <select
            id="category"
            value={selectedCategory}
            onChange={(e) => setSelectedCategory(e.target.value)}
            required
          >
            <option value="">Select a category</option>
            {Categories.map((category) => (
              <option key={category.title} value={category.title}>
                {category.title}
              </option>
            ))}
          </select>
        </div>
        <button type="submit">Submit Application</button>
      </form>
    </div>
  );
}

export default BecomeBusiness;
