import React, { useState, ChangeEvent, useEffect } from 'react';
import { useBusinessStore } from '../../../../../shared/stores/businessStore';
import {
  useUpdateBusiness,
  useSetBusinessSlug,
} from '../../../../../shared/hooks/queries/useBusiness';
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
  const { user } = useUserStore();
  const { websiteData, contactInfo } = useConfigStore();
  const updateBusiness = useUpdateBusiness();
  const setSlug = useSetBusinessSlug();
  const [slug, setSlugInput] = useState('');

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
        address: (websiteData?.address as string | undefined) || prev.address,
        city: (websiteData?.city as string | undefined) || prev.city,
        zipcode: (websiteData?.zipcode as string | undefined) || prev.zipcode,
        mapsUrl: (websiteData?.mapsUrl as string | undefined) || prev.mapsUrl,
        website: (websiteData?.website as string | undefined) || prev.website,
        country: (websiteData?.country as string | undefined) || prev.country,
        senderEmail:
          (websiteData?.senderEmail as string | undefined) ||
          contactInfo?.contactEmail ||
          prev.senderEmail,
      }));
    }
  }, [user, websiteData, contactInfo]);

  useEffect(() => {
    const w = websiteData as { slug?: string } | undefined;
    if (w?.slug) setSlugInput(w.slug);
  }, [websiteData]);

  const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    updateBusiness.mutate({ data: { ...formData, businessType } });
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
              <Label>Public slug</Label>
              <div className="flex gap-2 items-center">
                <span
                  className="text-xs whitespace-nowrap"
                  style={{ color: 'var(--dash-text-muted)' }}
                >
                  menu.rtylr.com/
                </span>
                <Input
                  name="slug"
                  value={slug}
                  placeholder="your-business"
                  onChange={(e) =>
                    setSlugInput(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))
                  }
                />
                <Button
                  type="button"
                  className="shrink-0 whitespace-nowrap"
                  disabled={!slug || setSlug.isPending}
                  onClick={() => setSlug.mutate(slug)}
                >
                  {setSlug.isPending ? 'Saving...' : 'Save slug'}
                </Button>
              </div>
              <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                Lowercase letters, numbers, and hyphens. Up to 80 chars.
              </p>
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
            {updateBusiness.isPending ? 'Saving...' : 'Save changes'}
          </Button>
        </div>
      </form>
    </div>
  );
};

export default General;
