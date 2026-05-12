import React, { useState, ChangeEvent, useEffect } from 'react';
import { useBusinessStore, BusinessType } from '../../../../../shared/stores/businessStore';
import { useUpdateBusiness } from '../../../../../shared/hooks/queries/useBusiness';
import { useUserStore } from '../../../../../shared/stores/userStore';
import { useConfigStore } from '../../../../../shared/stores/configStore';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  Input,
  Label,
  Button,
  Select,
} from '../../../../../shared/components/ui';

const BUSINESS_TYPES: { value: BusinessType; label: string }[] = [
  { value: 'restaurant', label: 'Restaurant' },
  { value: 'gym', label: 'Gym/Fitness' },
  { value: 'retail', label: 'Retail Store' },
  { value: 'salon', label: 'Salon/Beauty' },
  { value: 'cafe', label: 'Cafe' },
  { value: 'bar', label: 'Bar' },
  { value: 'hotel', label: 'Hotel' },
  { value: 'other', label: 'Other' },
];

const General: React.FC = () => {
  const { businessType, setBusinessType } = useBusinessStore();
  const { user, userId } = useUserStore();
  const { websiteData, contactInfo } = useConfigStore();
  const updateBusiness = useUpdateBusiness();

  const [formData, setFormData] = useState({
    storeName: '',
    contactEmail: '',
    senderEmail: '',
    storeIndustry: businessType || 'restaurant',
    cuisineType: '',
    companyName: '',
    website: '',
    telephone: '',
    mapsUrl: '',
    address: '',
    city: '',
    zipcode: '',
    country: 'Lebanon',
    state: '',
    customBusinessType: '',
  });

  useEffect(() => {
    if (user) {
      setFormData((prev) => ({
        ...prev,
        storeName:
          (websiteData?.siteName as string | undefined) || user.businessName || prev.storeName,
        contactEmail:
          contactInfo?.contactEmail ||
          (websiteData?.contactEmail as string | undefined) ||
          user.email ||
          prev.contactEmail,
        telephone:
          contactInfo?.contactNumber ||
          (websiteData?.contactNumber as string | undefined) ||
          prev.telephone,
        companyName:
          (websiteData?.siteName as string | undefined) || user.businessName || prev.companyName,
      }));
    }
  }, [user, websiteData, contactInfo]);

  useEffect(() => {
    if (businessType) setFormData((prev) => ({ ...prev, storeIndustry: businessType }));
  }, [businessType]);

  const handleInputChange = (e: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    if (name === 'storeIndustry') setBusinessType(value as BusinessType);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!userId) return;
    updateBusiness.mutate({ businessId: userId, data: formData });
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
          General
        </h2>
        <p className="text-sm mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
          Manage your store details and business information.
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Business Details */}
        <Card>
          <CardHeader>
            <CardTitle>Business details</CardTitle>
            <CardDescription>Your store profile information</CardDescription>
          </CardHeader>
          <CardContent className="pt-0 space-y-4">
            <div className="space-y-1.5">
              <Label>Store name</Label>
              <Input name="storeName" value={formData.storeName} onChange={handleInputChange} />
            </div>
            <div className="space-y-1.5">
              <Label>Contact email</Label>
              <Input
                type="email"
                name="contactEmail"
                value={formData.contactEmail}
                onChange={handleInputChange}
              />
            </div>
            <div className="space-y-1.5">
              <Label>Sender email</Label>
              <Input
                type="email"
                name="senderEmail"
                value={formData.senderEmail}
                onChange={handleInputChange}
              />
            </div>
            <div className="space-y-1.5">
              <Label>Store industry</Label>
              <Select
                name="storeIndustry"
                value={formData.storeIndustry}
                onChange={handleInputChange}
              >
                {BUSINESS_TYPES.map((type) => (
                  <option key={type.value} value={type.value}>
                    {type.label}
                  </option>
                ))}
              </Select>
            </div>
            {formData.storeIndustry === 'other' && (
              <div className="space-y-1.5">
                <Label>Custom business type</Label>
                <Input
                  name="customBusinessType"
                  value={formData.customBusinessType}
                  onChange={handleInputChange}
                  placeholder="Enter your business type"
                />
              </div>
            )}
            {formData.storeIndustry === 'restaurant' && (
              <div className="space-y-1.5">
                <Label>Cuisine type</Label>
                <Input
                  name="cuisineType"
                  value={formData.cuisineType}
                  onChange={handleInputChange}
                  placeholder="e.g. Italian, Mexican, Asian"
                />
              </div>
            )}
          </CardContent>
        </Card>

        {/* Business Address */}
        <Card>
          <CardHeader>
            <CardTitle>Business address</CardTitle>
            <CardDescription>Your physical business location</CardDescription>
          </CardHeader>
          <CardContent className="pt-0 space-y-4">
            {[
              { label: 'Company Name', name: 'companyName' },
              { label: 'Website', name: 'website' },
              { label: 'Telephone', name: 'telephone' },
              { label: 'Maps URL', name: 'mapsUrl' },
              { label: 'Address', name: 'address' },
              { label: 'City', name: 'city' },
              { label: 'Zipcode', name: 'zipcode' },
              { label: 'State', name: 'state' },
            ].map((field) => (
              <div key={field.name} className="space-y-1.5">
                <Label>{field.label}</Label>
                <Input
                  name={field.name}
                  value={(formData as any)[field.name]}
                  onChange={handleInputChange}
                />
              </div>
            ))}
            <div className="space-y-1.5">
              <Label>Country</Label>
              <Select name="country" value={formData.country} onChange={handleInputChange}>
                <option value="Lebanon">Lebanon</option>
              </Select>
            </div>
          </CardContent>
        </Card>

        <div className="flex justify-end">
          <Button type="submit" disabled={updateBusiness.isPending}>
            {updateBusiness.isPending ? 'Saving...' : 'Save Changes'}
          </Button>
        </div>
      </form>
    </div>
  );
};

export default General;
