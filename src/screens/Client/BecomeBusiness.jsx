import { useEffect, useState } from "react";
import { db, auth } from "../../firebase";

function BecomeBusiness() {
  // Define state variables to store form data
  const [requestStatus, setRequestStatus] = useState("pending"); // Can be "pending", "approved", or "denied"
  const [businessName, setBusinessName] = useState("");
  const [businessNameArabic, setBusinessNameArabic] = useState("");
  const [note, setNote] = useState("");
  const [country, setCountry] = useState(""); // New state for selected country
  const [requestExists, setRequestExists] = useState(false); // New state to track if a request exists
  const [isSubmitting, setIsSubmitting] = useState(false); // New state to track submission status

  useEffect(() => {
    // Check if a request exists for the current user
    db.collection("Requests")
      .doc(auth.currentUser.uid)
      .get()
      .then((doc) => {
        if (doc.exists) {
          setRequestStatus(doc.data().status);
          setRequestExists(true);
        }
      });
  }, []);

  // Handle form submission
  const handleSubmit = (e) => {
    e.preventDefault();
    if (!isSubmitting) {
      // Prevent multiple submissions
      setIsSubmitting(true);

      db.collection("Requests")
        .doc(auth.currentUser.uid)
        .set({
          businessName,
          businessNameArabic,
          note,
          createdAt: new Date(),
          userId: auth.currentUser?.uid,
          email: auth.currentUser?.email,
          status: "pending",
          country, // Include the selected country in the request
        })
        .then(() => {
          // Submission successful, update state
          setRequestStatus("pending");
        })
        .catch((error) => {
          // Handle submission error
          console.error("Error submitting request:", error);
        });
    }
  };

  return (
    <div className="BecomeBusiness-container">
      <h2>Become a Business</h2>
      {requestExists ? (
        requestStatus === "pending" ? (
          <div className="alert alert-warning">
            <strong>Request Pending!</strong> Your request is still pending.
          </div>
        ) : (
          <div className="alert alert-danger">
            <strong>Sorry!</strong> Your request has been denied.
          </div>
        )
      ) : (
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
              onChange={(e) => setBusinessNameArabic(e.target.value)}
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="note">Additional Note:</label>
            <textarea
              type="text"
              id="note"
              value={note}
              onChange={(e) => setNote(e.target.value)}
              required
              rows={6}
            />
          </div>
          <div className="form-group">
            <label htmlFor="country">Country:</label>
            <select
              id="country"
              value={country}
              onChange={(e) => setCountry(e.target.value)}
              required
            >
              <option value="">Select a Country</option>
              <option value="Bahrain">Bahrain</option>
              <option value="Kuwait">Kuwait</option>
              <option value="Oman">Oman</option>
              <option value="Qatar">Qatar</option>
              <option value="Saudi Arabia">Saudi Arabia</option>
              <option value="United Arab Emirates">United Arab Emirates</option>
            </select>
          </div>
          <button type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Request Pending" : "Submit Application"}
          </button>
        </form>
      )}
    </div>
  );
}

export default BecomeBusiness;
