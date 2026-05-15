import React, { useState, ChangeEvent, useEffect } from 'react';
import { useBusinessStore } from '../../../../../shared/stores/businessStore';
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
} from '../../../../../shared/components/ui';

const General: React.FC = () => {
  const { businessType } = useBusinessStore();
  const { user, userId } = useUserStore();
  const { websiteData, contactInfo } = useConfigStore();
  const updateBusiness = useUpdateBusiness();

  const [formData, setFormData] = useState({
    storeName: '',
    contactEmail: '',
    senderEmail: '',
    companyName: '',
    website: '',
    telephone: '',
    mapsUrl: '',
    address: '',
    city: '',
    zipcode: '',
    country: 'Lebanon',
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

  const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!userId) return;
    updateBusiness.mutate({
      businessId: userId,
      data: { ...formData, businessType },
    });
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
              <Input name="country" value={formData.country} onChange={handleInputChange} />
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
