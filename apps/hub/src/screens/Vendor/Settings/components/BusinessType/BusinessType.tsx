import React, { useState, useEffect } from "react";
import {
  useBusinessStore,
  BusinessType,
} from "../../../../../stores/businessStore";
import { useUpdateBusiness } from "../../../../../hooks/queries/useBusiness";
import { useUserStore } from "../../../../../stores/userStore";
import { Buildings as Building2 } from "phosphor-react";
import styles from "./BusinessType.module.scss";

const BUSINESS_TYPES: {
  value: BusinessType;
  label: string;
  description: string;
}[] = [
  {
    value: "restaurant",
    label: "Restaurant",
    description: "Food service and dining",
  },
  {
    value: "gym",
    label: "Gym/Fitness",
    description: "Fitness centers and gyms",
  },
  {
    value: "retail",
    label: "Retail Store",
    description: "Physical retail stores",
  },
  {
    value: "salon",
    label: "Salon/Beauty",
    description: "Beauty and personal care services",
  },
  { value: "cafe", label: "Cafe", description: "Coffee shops and cafes" },
  { value: "bar", label: "Bar", description: "Bars and nightlife" },
  { value: "hotel", label: "Hotel", description: "Hotels and accommodations" },
  { value: "other", label: "Other", description: "Other business types" },
];

const BusinessTypeSettings: React.FC = () => {
  const { businessType, setBusinessType } = useBusinessStore();
  const { userId } = useUserStore();
  const updateBusiness = useUpdateBusiness();
  const [selectedType, setSelectedType] = useState<BusinessType | null>(
    businessType
  );
  const [customType, setCustomType] = useState("");

  useEffect(() => {
    setSelectedType(businessType);
  }, [businessType]);

  const handleTypeChange = (type: BusinessType) => {
    setSelectedType(type);
    setBusinessType(type);
  };

  const handleSave = async () => {
    if (!userId || !selectedType) return;

    const data: any = {
      businessType: selectedType,
    };

    if (selectedType === "other" && customType) {
      data.customBusinessType = customType;
    }

    updateBusiness.mutate({
      businessId: userId,
      data,
    });
  };

  return (
    <div className={styles.businessType}>
      <div className={styles.header}>
        <div className={styles.headerIcon}>
          <Building2 size={24} />
        </div>
        <div>
          <h1>Business Type</h1>
          <p>
            Select the type of business you operate. This will customize your
            dashboard and features.
          </p>
        </div>
      </div>

      <div className={styles.businessTypesGrid}>
        {BUSINESS_TYPES.map((type) => (
          <div
            key={type.value}
            className={`${styles.businessTypeCard} ${
              selectedType === type.value ? styles.selected : ""
            }`}
            onClick={() => handleTypeChange(type.value)}
          >
            <div className={styles.cardHeader}>
              <input
                type="radio"
                name="businessType"
                value={type.value}
                checked={selectedType === type.value}
                onChange={() => handleTypeChange(type.value)}
              />
              <h3>{type.label}</h3>
            </div>
            <p>{type.description}</p>
          </div>
        ))}
      </div>

      {selectedType === "other" && (
        <div className={styles.customTypeSection}>
          <label htmlFor="customType">Custom Business Type</label>
          <input
            id="customType"
            type="text"
            value={customType}
            onChange={(e) => setCustomType(e.target.value)}
            placeholder="Enter your business type"
            className={styles.customInput}
          />
        </div>
      )}

      <div className={styles.actions}>
        <button
          onClick={handleSave}
          disabled={updateBusiness.isPending || !selectedType}
          className={styles.saveButton}
        >
          {updateBusiness.isPending ? "Saving..." : "Save Changes"}
        </button>
      </div>
    </div>
  );
};

export default BusinessTypeSettings;
