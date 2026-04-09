import React, { useState, ChangeEvent, useEffect } from "react";
import { useBusinessStore, BusinessType } from "../../../../../stores/businessStore";
import { useUpdateBusiness } from "../../../../../hooks/queries/useBusiness";
import { useUserStore } from "../../../../../stores/userStore";

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
    storeName: "", contactEmail: "", senderEmail: "",
    storeIndustry: businessType || "restaurant", cuisineType: "",
    companyName: "", website: "", telephone: "", mapsUrl: "",
    address: "", city: "", zipcode: "", country: "Lebanon", state: "",
    customBusinessType: "",
  });

  useEffect(() => {
    if (businessType) setFormData((prev) => ({ ...prev, storeIndustry: businessType }));
  }, [businessType]);

  const handleInputChange = (e: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    if (name === "storeIndustry") setBusinessType(value as BusinessType);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!userId) return;
    updateBusiness.mutate({ businessId: userId, data: formData });
  };

  const inputCls = "w-full border border-black/10 rounded-lg px-3 py-2 text-sm text-primary bg-white focus:outline-none focus:border-accent transition-colors";
  const labelCls = "block text-sm font-medium text-primary mb-1.5";

  return (
    <div className="max-w-2xl space-y-8">
      <div>
        <h2 className="text-xl font-bold text-primary">General</h2>
        <p className="text-sm text-muted mt-1">View and update your store details and business information.</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Business details */}
        <div className="bg-white rounded-xl border border-black/8 p-6 space-y-4">
          <h3 className="text-base font-semibold text-primary">Business details</h3>
          <div>
            <label className={labelCls}>Store name</label>
            <input type="text" name="storeName" value={formData.storeName} onChange={handleInputChange} className={inputCls} />
          </div>
          <div>
            <label className={labelCls}>Store contact email</label>
            <input type="email" name="contactEmail" value={formData.contactEmail} onChange={handleInputChange} className={inputCls} />
          </div>
          <div>
            <label className={labelCls}>Sender email</label>
            <input type="email" name="senderEmail" value={formData.senderEmail} onChange={handleInputChange} className={inputCls} />
          </div>
          <div>
            <label className={labelCls}>Store industry</label>
            <select name="storeIndustry" value={formData.storeIndustry} onChange={handleInputChange} className={inputCls}>
              {BUSINESS_TYPES.map((type) => (
                <option key={type.value} value={type.value}>{type.label}</option>
              ))}
            </select>
          </div>
          {formData.storeIndustry === "other" && (
            <div>
              <label className={labelCls}>Custom business type</label>
              <input type="text" name="customBusinessType" value={formData.customBusinessType} onChange={handleInputChange} placeholder="Enter your business type" className={inputCls} />
            </div>
          )}
          {formData.storeIndustry === "restaurant" && (
            <div>
              <label className={labelCls}>Cuisine type</label>
              <input type="text" name="cuisineType" value={formData.cuisineType} onChange={handleInputChange} placeholder="e.g., Italian, Mexican, Asian" className={inputCls} />
            </div>
          )}
        </div>

        {/* Business address */}
        <div className="bg-white rounded-xl border border-black/8 p-6 space-y-4">
          <h3 className="text-base font-semibold text-primary">Business address</h3>
          {[
            { label: "Company Name", name: "companyName" },
            { label: "Website", name: "website" },
            { label: "Telephone", name: "telephone" },
            { label: "Maps URL", name: "mapsUrl" },
            { label: "Address", name: "address" },
            { label: "City", name: "city" },
            { label: "Zipcode", name: "zipcode" },
            { label: "State", name: "state" },
          ].map((field) => (
            <div key={field.name}>
              <label className={labelCls}>{field.label}</label>
              <input type="text" name={field.name} value={(formData as any)[field.name]} onChange={handleInputChange} className={inputCls} />
            </div>
          ))}
          <div>
            <label className={labelCls}>Country</label>
            <select name="country" value={formData.country} onChange={handleInputChange} className={inputCls}>
              <option value="Lebanon">Lebanon</option>
            </select>
          </div>
        </div>

        <div className="flex justify-end">
          <button type="submit" disabled={updateBusiness.isPending} className="px-5 py-2 text-sm font-medium rounded-lg bg-admin text-white hover:opacity-90 transition-opacity disabled:opacity-60">
            {updateBusiness.isPending ? "Saving..." : "Save Changes"}
          </button>
        </div>
      </form>
    </div>
  );
};

export default General;
