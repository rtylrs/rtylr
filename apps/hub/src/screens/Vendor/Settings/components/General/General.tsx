import React, { useState, ChangeEvent, useEffect } from "react";
import {
  useBusinessStore,
  BusinessType,
} from "../../../../../stores/businessStore";
import { useUpdateBusiness } from "../../../../../hooks/queries/useBusiness";
import { useUserStore } from "../../../../../stores/userStore";
import styles from "./General.module.scss";

const BUSINESS_TYPES: { value: BusinessType; label: string }[] = [
  { value: "restaurant", label: "Restaurant" },
  { value: "gym", label: "Gym/Fitness" },
  { value: "retail", label: "Retail Store" },
  { value: "salon", label: "Salon/Beauty" },
  { value: "cafe", label: "Cafe" },
  { value: "bar", label: "Bar" },
  { value: "hotel", label: "Hotel" },
  { value: "other", label: "Other" },
];

const General: React.FC = () => {
  const { businessType, setBusinessType } = useBusinessStore();
  const { userId } = useUserStore();
  const updateBusiness = useUpdateBusiness();

  const [formData, setFormData] = useState({
    storeName: "",
    contactEmail: "",
    senderEmail: "",
    storeIndustry: businessType || "restaurant",
    cuisineType: "",
    companyName: "",
    website: "",
    telephone: "",
    mapsUrl: "",
    address: "",
    city: "",
    zipcode: "",
    country: "Lebanon",
    state: "",
    customBusinessType: "",
  });

  useEffect(() => {
    if (businessType) {
      setFormData((prev) => ({ ...prev, storeIndustry: businessType }));
    }
  }, [businessType]);

  const handleInputChange = (
    e: ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target;
    setFormData((prevData) => ({
      ...prevData,
      [name]: value,
    }));

    if (name === "storeIndustry") {
      setBusinessType(value as BusinessType);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!userId) return;

    updateBusiness.mutate({
      businessId: userId,
      data: formData,
    });
  };

  return (
    <div className={styles.general}>
      <div className={styles.header}>
        <h2>General</h2>
        <p>View and update your store details and business information.</p>
      </div>
      <form onSubmit={handleSubmit}>
        <div className={styles.formSection}>
          <h3>Business details</h3>
          <div className={styles.formGroup}>
            <label>Store name</label>
            <input
              type="text"
              name="storeName"
              value={formData.storeName}
              onChange={handleInputChange}
            />
          </div>
          <div className={styles.formGroup}>
            <label>Store contact email</label>
            <input
              type="email"
              name="contactEmail"
              value={formData.contactEmail}
              onChange={handleInputChange}
            />
          </div>
          <div className={styles.formGroup}>
            <label>Sender email</label>
            <input
              type="email"
              name="senderEmail"
              value={formData.senderEmail}
              onChange={handleInputChange}
            />
          </div>
          <div className={styles.formGroup}>
            <label>Store industry</label>
            <select
              name="storeIndustry"
              value={formData.storeIndustry}
              onChange={handleInputChange}
            >
              {BUSINESS_TYPES.map((type) => (
                <option key={type.value} value={type.value}>
                  {type.label}
                </option>
              ))}
            </select>
          </div>
          {formData.storeIndustry === "other" && (
            <div className={styles.formGroup}>
              <label>Custom business type</label>
              <input
                type="text"
                name="customBusinessType"
                value={formData.customBusinessType}
                onChange={handleInputChange}
                placeholder="Enter your business type"
              />
            </div>
          )}
          {formData.storeIndustry === "restaurant" && (
            <div className={styles.formGroup}>
              <label>Cuisine type</label>
              <input
                type="text"
                name="cuisineType"
                value={formData.cuisineType}
                onChange={handleInputChange}
                placeholder="e.g., Italian, Mexican, Asian"
              />
            </div>
          )}
        </div>
        <div className={styles.formSection}>
          <h3>Business address</h3>
          <div className={styles.formGroup}>
            <label>Company Name</label>
            <input
              type="text"
              name="companyName"
              value={formData.companyName}
              onChange={handleInputChange}
            />
          </div>
          <div className={styles.formGroup}>
            <label>Website</label>
            <input
              type="text"
              name="website"
              value={formData.website}
              onChange={handleInputChange}
            />
          </div>
          <div className={styles.formGroup}>
            <label>Telephone</label>
            <input
              type="text"
              name="telephone"
              value={formData.telephone}
              onChange={handleInputChange}
            />
          </div>
          <div className={styles.formGroup}>
            <label>Maps url</label>
            <input
              type="text"
              name="mapsUrl"
              value={formData.mapsUrl}
              onChange={handleInputChange}
            />
          </div>
          <div className={styles.formGroup}>
            <label>Address</label>
            <input
              type="text"
              name="address"
              value={formData.address}
              onChange={handleInputChange}
            />
          </div>
          <div className={styles.formGroup}>
            <label>City</label>
            <input
              type="text"
              name="city"
              value={formData.city}
              onChange={handleInputChange}
            />
          </div>
          <div className={styles.formGroup}>
            <label>Zipcode</label>
            <input
              type="text"
              name="zipcode"
              value={formData.zipcode}
              onChange={handleInputChange}
            />
          </div>
          <div className={styles.formGroup}>
            <label>Country</label>
            <select
              name="country"
              value={formData.country}
              onChange={handleInputChange}
            >
              <option value="Lebanon">Lebanon</option>
            </select>
          </div>
          <div className={styles.formGroup}>
            <label>State</label>
            <input
              type="text"
              name="state"
              value={formData.state}
              onChange={handleInputChange}
            />
          </div>
        </div>
        <div className={styles.formActions}>
          <button type="submit" disabled={updateBusiness.isPending}>
            {updateBusiness.isPending ? "Saving..." : "Save Changes"}
          </button>
        </div>
      </form>
    </div>
  );
};

export default General;
