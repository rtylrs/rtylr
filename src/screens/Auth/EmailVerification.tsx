import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { CircularProgress } from "@mui/material";
import api from "../../lib/api";

const EmailVerification = () => {
  const { token } = useParams();
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const verifyEmail = async () => {
      try {
        const response = await api.get(`/auth/verify/${token}`);
        setMessage(response.data.message);
        setLoading(false);
      } catch (error) {
        setMessage("Email verification failed. Please try again.");
        setLoading(false);
      }
    };

    verifyEmail();
  }, [token]);

  return (
    <div style={styles.container}>
      <h1>Email Verification</h1>
      {loading ? <CircularProgress /> : <p>{message}</p>}
    </div>
  );
};

const styles = {
  container: {
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
    justifyContent: "center",
    height: "60vh",
    backgroundColor: "#f5f5f5",
    textAlign: "center",
  },
};

export default EmailVerification;
