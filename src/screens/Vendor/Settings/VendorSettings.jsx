import React from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faCog,
  faUserShield,
  faCreditCard,
  faFile,
  faDatabase,
  faLanguage,
  faGavel,
} from "@fortawesome/free-solid-svg-icons";

const VendorSettings = () => {
  return (
    <div className="vendor-settings-container">
      <h1>Settings</h1>
      <div className="settings-grid">
        <div className="settings-item">
          <div className="settings-icon">
            <FontAwesomeIcon icon={faCog} />
          </div>
          <div className="settings-content">
            <h2>General</h2>
            <p>View and update your store details</p>
          </div>
        </div>
        <div className="settings-item">
          <div className="settings-icon">
            <FontAwesomeIcon icon={faUserShield} />
          </div>
          <div className="settings-content">
            <h2>Plan and permissions</h2>
            <p>
              View plan information and manage what staff can see or do in your
              store
            </p>
          </div>
        </div>
        <div className="settings-item">
          <div className="settings-icon">
            <FontAwesomeIcon icon={faCreditCard} />
          </div>
          <div className="settings-content">
            <h2>Payments</h2>
            <p>Enable and manage your store's payment providers</p>
          </div>
        </div>
        <div className="settings-item">
          <div className="settings-icon">
            <FontAwesomeIcon icon={faFile} />
          </div>
          <div className="settings-content">
            <h2>Files</h2>
            <p>Upload images, videos, and documents</p>
          </div>
        </div>
        <div className="settings-item">
          <div className="settings-icon">
            <FontAwesomeIcon icon={faDatabase} />
          </div>
          <div className="settings-content">
            <h2>Backups</h2>
            <p>Check your last backups</p>
          </div>
        </div>
        <div className="settings-item">
          <div className="settings-icon">
            <FontAwesomeIcon icon={faLanguage} />
          </div>
          <div className="settings-content">
            <h2>Store languages</h2>
            <p>Manage the languages your customers can view on your store</p>
          </div>
        </div>
        <div className="settings-item">
          <div className="settings-icon">
            <FontAwesomeIcon icon={faGavel} />
          </div>
          <div className="settings-content">
            <h2>Legal</h2>
            <p>Manage your store's legal pages</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default VendorSettings;
