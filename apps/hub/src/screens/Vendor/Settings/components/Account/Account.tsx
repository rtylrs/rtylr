import React, { useState, useEffect } from "react";
import { useUserStore } from "../../../../../stores/userStore";
import styles from "./Account.module.scss";

interface LoginHistoryItem {
  email: string;
  method: string;
  ip: string;
  timeAgo: string;
}

const Account: React.FC = () => {
  const { user } = useUserStore();
  const [formData, setFormData] = useState({
    firstName: "",
    lastName: "",
    email: "",
  });
  const [isEditing, setIsEditing] = useState(false);
  const [loginHistory, setLoginHistory] = useState<LoginHistoryItem[]>([
    {
      email: "mohamad.homsi@anghami.com",
      method: "Logged in via Google",
      ip: "from 141.105.87.106",
      timeAgo: "3 days ago",
    },
    {
      email: "mohamad.homsi@anghami.com",
      method: "Logged in via Google",
      ip: "from 141.105.87.106",
      timeAgo: "2 weeks ago",
    },
    {
      email: "mohamad.homsi@anghami.com",
      method: "Logged in via Google",
      ip: "from 141.105.87.106",
      timeAgo: "2 weeks ago",
    },
  ]);
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 3;

  useEffect(() => {
    if (user) {
      setFormData({
        firstName: user.firstName || user.name?.split(" ")[0] || "",
        lastName: user.lastName || user.name?.split(" ").slice(1).join(" ") || "",
        email: user.email || "",
      });
    }
  }, [user]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    setIsEditing(true);
  };

  const handleSave = () => {
    // TODO: Implement save logic
    console.log("Saving:", formData);
    setIsEditing(false);
  };

  const handleCancel = () => {
    if (user) {
      setFormData({
        firstName: user.firstName || user.name?.split(" ")[0] || "",
        lastName: user.lastName || user.name?.split(" ").slice(1).join(" ") || "",
        email: user.email || "",
      });
    }
    setIsEditing(false);
  };

  const handleInvalidateSessions = () => {
    // TODO: Implement invalidate sessions
    console.log("Invalidate all sessions");
  };

  const totalPages = Math.ceil(loginHistory.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const currentHistory = loginHistory.slice(startIndex, endIndex);

  return (
    <div className={styles.account}>
      <div className={styles.header}>
        <h1 className={styles.title}>Account</h1>
      </div>

      <div className={styles.section}>
        <div className={styles.sectionContent}>
          <div className={styles.formGroup}>
            <label htmlFor="firstName">First name</label>
            <input
              type="text"
              id="firstName"
              name="firstName"
              value={formData.firstName}
              onChange={handleInputChange}
              className={styles.input}
            />
          </div>
          <div className={styles.formGroup}>
            <label htmlFor="lastName">Last name</label>
            <input
              type="text"
              id="lastName"
              name="lastName"
              value={formData.lastName}
              onChange={handleInputChange}
              className={styles.input}
            />
          </div>
          <div className={styles.formGroup}>
            <label htmlFor="email">Email</label>
            <input
              type="email"
              id="email"
              name="email"
              value={formData.email}
              onChange={handleInputChange}
              className={styles.input}
            />
          </div>
        </div>
        <div className={styles.actions}>
          <button
            type="button"
            onClick={handleCancel}
            className={styles.cancelButton}
            disabled={!isEditing}
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleSave}
            className={styles.saveButton}
            disabled={!isEditing}
          >
            Save changes
          </button>
        </div>
      </div>

      <div className={styles.section}>
        <h2 className={styles.sectionTitle}>Invalidate sessions</h2>
        <p className={styles.description}>
          Invalidate all sessions for your account. This will log you out of your
          current session.
        </p>
        <button
          type="button"
          onClick={handleInvalidateSessions}
          className={styles.dangerButton}
        >
          Invalidate all account sessions
        </button>
      </div>

      <div className={styles.section}>
        <h2 className={styles.sectionTitle}>Login history</h2>
        <div className={styles.tableContainer}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th>Email</th>
                <th>Method</th>
                <th>IP Address</th>
                <th>Time</th>
              </tr>
            </thead>
            <tbody>
              {currentHistory.map((item, index) => (
                <tr key={index}>
                  <td>{item.email}</td>
                  <td>{item.method}</td>
                  <td>{item.ip}</td>
                  <td>{item.timeAgo}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {totalPages > 1 && (
          <div className={styles.pagination}>
            <button
              type="button"
              onClick={() => setCurrentPage((prev) => Math.max(1, prev - 1))}
              disabled={currentPage === 1}
              className={styles.paginationButton}
            >
              ←
            </button>
            {Array.from({ length: totalPages }, (_, i) => i + 1).map((page) => (
              <button
                key={page}
                type="button"
                onClick={() => setCurrentPage(page)}
                className={`${styles.paginationButton} ${
                  currentPage === page ? styles.active : ""
                }`}
              >
                {page}
              </button>
            ))}
            <button
              type="button"
              onClick={() =>
                setCurrentPage((prev) => Math.min(totalPages, prev + 1))
              }
              disabled={currentPage === totalPages}
              className={styles.paginationButton}
            >
              →
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default Account;

