import React, { useRef, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { uploadAsset } from '@voxire/api';
import { Palette, Upload } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import { Button, Input, Label } from '../../shared/components/ui';
import { SettingSection } from '../../shared/components/dash';
import { useUpdateBusiness } from '../../shared/hooks/queries/useBusiness';
import { useConfigStore } from '../../shared/stores/configStore';

const COLORS = [
  '#4583C4',
  '#64C5B9',
  '#10b981',
  '#f59e0b',
  '#ef4444',
  '#8b5cf6',
  '#ec4899',
  '#06b6d4',
];

const ThemeCustomizer: React.FC = () => {
  const navigate = useNavigate();
  const updateBusiness = useUpdateBusiness();
  const websiteData = useConfigStore((s) => s.websiteData);

  const [accent, setAccent] = useState((websiteData?.primaryColor as string) || '#4583C4');
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
      console.error(err);
    } finally {
      setUploadingLogo(false);
      if (logoInputRef.current) logoInputRef.current.value = '';
    }
  };

  const handleSave = () => {
    document.documentElement.style.setProperty('--dash-accent', accent);
    updateBusiness.mutate({
      data: {
        primaryColor: accent,
        brandName,
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

        <SettingSection
          title="Accent color"
          description="Applied to primary buttons, links, and highlights."
          footer={
            <Button size="sm" onClick={() => navigate({ to: '/design' })}>
              Open full design editor
            </Button>
          }
        >
          <div className="space-y-4">
            <div className="flex items-center gap-2 flex-wrap">
              {COLORS.map((c) => (
                <button
                  type="button"
                  key={c}
                  onClick={() => setAccent(c)}
                  className="h-9 w-9 rounded-lg transition-all"
                  style={{
                    background: c,
                    border:
                      accent === c ? '2px solid var(--dash-text)' : '1px solid var(--dash-border)',
                    transform: accent === c ? 'scale(1.1)' : 'scale(1)',
                  }}
                  aria-label={`Select ${c}`}
                />
              ))}
            </div>
            <div
              className="flex items-center gap-3 px-4 py-3 rounded-lg"
              style={{ background: 'var(--dash-hover)' }}
            >
              <div className="h-10 w-10 rounded-lg" style={{ background: accent }} />
              <div className="flex-1">
                <p className="text-xs font-semibold" style={{ color: 'var(--dash-text)' }}>
                  Preview
                </p>
                <code className="text-[11px] font-mono" style={{ color: 'var(--dash-text-dim)' }}>
                  {accent}
                </code>
              </div>
              <button
                type="button"
                className="px-4 py-2 rounded-lg text-xs font-semibold"
                style={{ background: accent, color: 'white' }}
              >
                Sample button
              </button>
            </div>
          </div>
        </SettingSection>

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
