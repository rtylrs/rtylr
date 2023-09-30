import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { db, auth } from "../../firebase";
import i18next from "../../services/i18next";

const ProfilePage = () => {
  const { t } = useTranslation();
  const [editing, setEditing] = useState(false);
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    notifications: true,
    address: "",
    paymentInfo: "",
  });

  useEffect(() => {
    // Fetch user information from Firebase Authentication
    const user = auth.currentUser;
    if (user) {
      setFormData({
        name: user.displayName || "",
        email: user.email || "",
        notifications: true,
        address: "",
        paymentInfo: "",
      });
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleEditClick = (field) => {
    setEditing(field);
  };

  const handleSaveClick = () => {
    // Update user data in the backend (Firestore)
    const user = auth.currentUser;
    if (user) {
      // Update user's display name
      if (editing === "name") {
        user
          .updateProfile({
            displayName: formData.name,
          })
          .then(() => {
            setEditing(false);
          })
          .catch((error) => {
            console.error("Error updating display name:", error);
          });
      }
      // Update user's data in Firestore (e.g., address, payment info)
      else {
        const userRef = db.collection("Users").doc(user.uid);
        userRef
          .set({
            [editing]: formData[editing],
          })
          .then(() => {
            setEditing(false);
          })
          .catch((error) => {
            console.error("Error updating user data:", error);
          });
      }
    }
  };

  const handleCancelClick = () => {
    setEditing(false);
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value,
    });
  };

  return (
    <div className="profile-container">
      <h2>{t("Profile.title")}</h2>
      <div>
        <p>{t("Profile.name")}:</p>
        {editing === "name" ? (
          <input
            type="text"
            name="name"
            value={formData.name}
            onChange={handleChange}
          />
        ) : (
          <p>{formData.name}</p>
        )}
        {editing !== "name" && (
          <button
            className="edit-button"
            onClick={() => handleEditClick("name")}
          >
            {t("Profile.edit")}
          </button>
        )}
      </div>
      <div>
        <p>{t("Profile.email")}:</p>
        <p>{formData.email}</p>
      </div>
      <div>
        <p>{t("Profile.notifications")}:</p>
        <p>{formData.notifications ? "Enabled" : "Disabled"}</p>
      </div>
      <div>
        <p>{t("Profile.address")}:</p>
        {editing === "address" ? (
          <textarea
            name="address"
            value={formData.address}
            onChange={handleChange}
          />
        ) : (
          <p>{formData.address}</p>
        )}
        {editing !== "address" && (
          <button
            className="edit-button"
            onClick={() => handleEditClick("address")}
          >
            {t("Profile.edit")}
          </button>
        )}
      </div>
      {editing && (
        <div>
          <button className="save-button" onClick={handleSaveClick}>
            {t("Profile.save")}
          </button>
          <button className="cancel-button" onClick={handleCancelClick}>
            {t("Profile.cancel")}
          </button>
        </div>
      )}
    </div>
  );
};

export default ProfilePage;
