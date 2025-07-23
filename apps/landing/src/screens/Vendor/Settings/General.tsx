import React, { useState, ChangeEvent } from "react";

const General: React.FC = () => {
  const [formData, setFormData] = useState({
    storeName: "al-shami",
    contactEmail: "homsiomohamad17@gmail.com",
    senderEmail: "homsiomohamad17@gmail.com",
    storeIndustry: "Restaurant",
    cuisineType: "Falafel - Shawarma",
    companyName: "al shami",
    website: "https://dopweb.com/",
    telephone: "",
    mapsUrl: "",
    address: "",
    city: "Beirut",
    zipcode: "",
    country: "Lebanon",
    state: "",
  });

  const handleInputChange = (
    e: ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target;
    setFormData((prevData) => ({
      ...prevData,
      [name]: value,
    }));
  };

  return (
    <div className="general">
      <h2>General</h2>
      <div className="form-section">
        <h3>Business details</h3>
        <div className="form-group">
          <label>Store name</label>
          <input
            type="text"
            name="storeName"
            value={formData.storeName}
            onChange={handleInputChange}
          />
        </div>
        <div className="form-group">
          <label>Store contact email</label>
          <input
            type="email"
            name="contactEmail"
            value={formData.contactEmail}
            onChange={handleInputChange}
          />
        </div>
        <div className="form-group">
          <label>Sender email</label>
          <input
            type="email"
            name="senderEmail"
            value={formData.senderEmail}
            onChange={handleInputChange}
          />
        </div>
        <div className="form-group">
          <label>Store industry</label>
          <select
            name="storeIndustry"
            value={formData.storeIndustry}
            onChange={handleInputChange}
          >
            <option value="Restaurant">Restaurant</option>
          </select>
        </div>
        <div className="form-group">
          <label>Cuisine type</label>
          <input
            type="text"
            name="cuisineType"
            value={formData.cuisineType}
            onChange={handleInputChange}
          />
        </div>
      </div>
      <div className="form-section">
        <h3>Business address</h3>
        <div className="form-group">
          <label>Company Name</label>
          <input
            type="text"
            name="companyName"
            value={formData.companyName}
            onChange={handleInputChange}
          />
        </div>
        <div className="form-group">
          <label>Website</label>
          <input
            type="text"
            name="website"
            value={formData.website}
            onChange={handleInputChange}
          />
        </div>
        <div className="form-group">
          <label>Telephone</label>
          <input
            type="text"
            name="telephone"
            value={formData.telephone}
            onChange={handleInputChange}
          />
        </div>
        <div className="form-group">
          <label>Maps url</label>
          <input
            type="text"
            name="mapsUrl"
            value={formData.mapsUrl}
            onChange={handleInputChange}
          />
        </div>
        <div className="form-group">
          <label>Address</label>
          <input
            type="text"
            name="address"
            value={formData.address}
            onChange={handleInputChange}
          />
        </div>
        <div className="form-group">
          <label>City</label>
          <input
            type="text"
            name="city"
            value={formData.city}
            onChange={handleInputChange}
          />
        </div>
        <div className="form-group">
          <label>Zipcode</label>
          <input
            type="text"
            name="zipcode"
            value={formData.zipcode}
            onChange={handleInputChange}
          />
        </div>
        <div className="form-group">
          <label>Country</label>
          <select
            name="country"
            value={formData.country}
            onChange={handleInputChange}
          >
            <option value="Lebanon">Lebanon</option>
          </select>
        </div>
        <div className="form-group">
          <label>State</label>
          <input
            type="text"
            name="state"
            value={formData.state}
            onChange={handleInputChange}
          />
        </div>
      </div>
    </div>
  );
};

export default General;
