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
            <span>View and update your store details</span>
          </div>
        </div>
        <div className="settings-item">
          <div className="settings-icon">
            <FontAwesomeIcon icon={faUserShield} />
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
            <FontAwesomeIcon icon={faCreditCard} />
          </div>
          <div className="settings-content">
            <h2>Payments</h2>
            <span>Enable and manage your store's payment providers</span>
          </div>
        </div>
        <div className="settings-item">
          <div className="settings-icon">
            <FontAwesomeIcon icon={faFile} />
          </div>
          <div className="settings-content">
            <h2>Files</h2>
            <span>Upload images, videos, and documents</span>
          </div>
        </div>
        <div className="settings-item">
          <div className="settings-icon">
            <FontAwesomeIcon icon={faLanguage} />
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
            <FontAwesomeIcon icon={faGavel} />
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
