import React, { useState, useEffect } from "react";
import { useBusinessStore, BusinessType } from "../../../../../stores/businessStore";
import { useUpdateBusiness } from "../../../../../hooks/queries/useBusiness";
import { useUserStore } from "../../../../../stores/userStore";
import { Buildings as Building2 } from "phosphor-react";

const BUSINESS_TYPES: { value: BusinessType; label: string; description: string }[] = [
  { value: "restaurant", label: "Restaurant", description: "Food service and dining" },
  { value: "gym", label: "Gym/Fitness", description: "Fitness centers and gyms" },
  { value: "retail", label: "Retail Store", description: "Physical retail stores" },
  { value: "salon", label: "Salon/Beauty", description: "Beauty and personal care services" },
  { value: "cafe", label: "Cafe", description: "Coffee shops and cafes" },
  { value: "bar", label: "Bar", description: "Bars and nightlife" },
  { value: "hotel", label: "Hotel", description: "Hotels and accommodations" },
  { value: "other", label: "Other", description: "Other business types" },
];

const BusinessTypeSettings: React.FC = () => {
  const { businessType, setBusinessType } = useBusinessStore();
  const { userId } = useUserStore();
  const updateBusiness = useUpdateBusiness();
  const [selectedType, setSelectedType] = useState<BusinessType | null>(businessType);
  const [customType, setCustomType] = useState("");

  useEffect(() => { setSelectedType(businessType); }, [businessType]);

  const handleTypeChange = (type: BusinessType) => {
    setSelectedType(type);
    setBusinessType(type);
  };

  const handleSave = async () => {
    if (!userId || !selectedType) return;
    const data: any = { businessType: selectedType };
    if (selectedType === "other" && customType) data.customBusinessType = customType;
    updateBusiness.mutate({ businessId: userId, data });
  };

  return (
    <div className="max-w-2xl space-y-6">
      <div className="flex items-start gap-4">
        <div className="w-10 h-10 rounded-xl bg-admin/10 flex items-center justify-center text-admin shrink-0">
          <Building2 size={20} />
        </div>
        <div>
          <h1 className="text-xl font-bold text-primary">Business Type</h1>
          <p className="text-sm text-muted mt-0.5">Select the type of business you operate. This will customize your dashboard and features.</p>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3">
        {BUSINESS_TYPES.map((type) => (
          <div
            key={type.value}
            className={`p-4 rounded-xl border cursor-pointer transition-all ${
              selectedType === type.value
                ? "border-admin bg-admin/5 shadow-sm"
                : "border-black/8 bg-white hover:border-black/20"
            }`}
            onClick={() => handleTypeChange(type.value)}
          >
            <div className="flex items-center gap-2 mb-1">
              <input
                type="radio"
                name="businessType"
                value={type.value}
                checked={selectedType === type.value}
                onChange={() => handleTypeChange(type.value)}
                className="accent-admin"
              />
              <h3 className="text-sm font-semibold text-primary">{type.label}</h3>
            </div>
            <p className="text-xs text-muted pl-5">{type.description}</p>
          </div>
        ))}
      </div>

      {selectedType === "other" && (
        <div>
          <label className="block text-sm font-medium text-primary mb-1.5">Custom Business Type</label>
          <input
            id="customType"
            type="text"
            value={customType}
            onChange={(e) => setCustomType(e.target.value)}
            placeholder="Enter your business type"
            className="w-full border border-black/10 rounded-lg px-3 py-2 text-sm text-primary bg-white focus:outline-none focus:border-accent transition-colors"
          />
        </div>
      )}

      <div className="flex justify-end">
        <button
          onClick={handleSave}
          disabled={updateBusiness.isPending || !selectedType}
          className="px-5 py-2 text-sm font-medium rounded-lg bg-admin text-white hover:opacity-90 transition-opacity disabled:opacity-60"
        >
          {updateBusiness.isPending ? "Saving..." : "Save Changes"}
        </button>
      </div>
    </div>
  );
};

export default BusinessTypeSettings;
