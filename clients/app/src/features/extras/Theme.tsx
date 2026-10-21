import React, { useRef, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { uploadAsset, getOneKitErrorMessage } from '@voxire/api';
import { Palette, Upload } from 'lucide-react';
import { toast } from 'sonner';
import { Button, Input, Label } from '../../shared/components/ui';
import { SettingSection } from '../../shared/components/dash';
import { useUpdateBusiness } from '../../shared/hooks/queries/useBusiness';
import { useConfigStore } from '../../shared/stores/configStore';

const ThemeCustomizer: React.FC = () => {
  const updateBusiness = useUpdateBusiness();
  const websiteData = useConfigStore((s) => s.websiteData);

  const [brandName, setBrandName] = useState(
    (websiteData?.brandName as string) || (websiteData?.siteName as string) || 'My Business'
  );
  const [logoUrl, setLogoUrl] = useState((websiteData?.logoUrl as string) || '');
  const [uploadingLogo, setUploadingLogo] = useState(false);
  const logoInputRef = useRef<HTMLInputElement>(null);

  const handleLogoUpload = async (file?: File | null): Promise<void> => {
    if (!file) return;
    setUploadingLogo(true);
    try {
      const uploaded = await uploadAsset(file);
      setLogoUrl(uploaded.url);
    } catch (err) {
      toast.error(getOneKitErrorMessage(err, 'Logo upload failed'));
    } finally {
      setUploadingLogo(false);
      if (logoInputRef.current) logoInputRef.current.value = '';
    }
  };

  const handleSave = () => {
    // The config contract has no brand_name or primary_color fields, so only
    // the fields it can carry are sent. The business name maps to siteName.
    updateBusiness.mutate({
      data: {
        storeName: brandName,
        logoUrl,
      },
    });
  };

  return (
    <>
      <HelmetContainer title="Theme · rtylr" />
      <div className="space-y-5 max-w-3xl">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div
              className="h-9 w-9 rounded-xl flex items-center justify-center"
              style={{ background: 'rgba(69,131,196,0.15)' }}
            >
              <Palette size={16} style={{ color: '#4583C4' }} />
            </div>
            <div>
              <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
                Theme
              </h1>
              <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                Your workspace branding.
              </p>
            </div>
          </div>
          <Button onClick={handleSave} disabled={updateBusiness.isPending}>
            {updateBusiness.isPending ? 'Saving...' : 'Save Theme'}
          </Button>
        </div>

        <SettingSection title="Brand" description="Your business name and logo.">
          <div className="space-y-3">
            <div className="space-y-1.5">
              <Label>Business name</Label>
              <Input value={brandName} onChange={(e) => setBrandName(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label>Logo URL</Label>
              <div className="flex gap-2">
                <Input
                  value={logoUrl}
                  onChange={(e) => setLogoUrl(e.target.value)}
                  placeholder="https://..."
                />
                <Button
                  variant="secondary"
                  disabled={uploadingLogo}
                  onClick={() => logoInputRef.current?.click()}
                >
                  <Upload size={13} />
                  {uploadingLogo ? 'Uploading...' : 'Upload'}
                </Button>
                <input
                  ref={logoInputRef}
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={(e) => {
                    void handleLogoUpload(e.target.files?.[0]);
                  }}
                />
              </div>
            </div>
          </div>
        </SettingSection>
      </div>
    </>
  );
};

export default ThemeCustomizer;
