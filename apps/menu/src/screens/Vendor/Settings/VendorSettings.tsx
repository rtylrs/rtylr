import {
  FaCog,
  FaUserShield,
  FaCreditCard,
  FaFile,
  FaLanguage,
  FaGavel,
} from "react-icons/fa";
import { useNavigate } from "react-router-dom";

const VendorSettings = () => {
  const navigate = useNavigate();

  return (
    <div className="vendor-settings-container">
      <h1>Settings</h1>
      <div className="settings-grid">
        <div className="settings-item">
          <div className="settings-icon">
            <FaCog />
          </div>
          <div
            className="settings-content"
            onClick={() => {
              navigate("/app/business/settings/general");
            }}
          >
            <h2>General</h2>
            <span>View and update your store details</span>
          </div>
        </div>
        <div className="settings-item">
          <div className="settings-icon">
            <FaUserShield />
          </div>
          <div className="settings-content">
            <h2>Plan and permissions</h2>
            <span>
              View plan information and manage what staff can see or do in your
              store
            </span>
          </div>
        </div>
        <div className="settings-item">
          <div className="settings-icon">
            <FaCreditCard />
          </div>
          <div className="settings-content">
            <h2>Payments</h2>
            <span>Enable and manage your store's payment providers</span>
          </div>
        </div>
        <div className="settings-item">
          <div className="settings-icon">
            <FaFile />
          </div>
          <div className="settings-content">
            <h2>Files</h2>
            <span>Upload images, videos, and documents</span>
          </div>
        </div>
        <div className="settings-item">
          <div className="settings-icon">
            <FaLanguage />
          </div>
          <div className="settings-content">
            <h2>Store languages</h2>
            <span>
              Manage the languages your customers can view on your store
            </span>
          </div>
        </div>
        <div className="settings-item">
          <div className="settings-icon">
            <FaGavel />
          </div>
          <div className="settings-content">
            <h2>Legal</h2>
            <span>Manage your store's legal pages</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default VendorSettings;
