import React, { useState, ChangeEvent, useEffect, useMemo } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { AnimatePresence, motion } from 'motion/react';
import { Check } from 'lucide-react';
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
import { cn } from '../../../../../shared/lib/utils';

const INITIAL_FORM_DATA = {
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
};

type BusinessFormData = typeof INITIAL_FORM_DATA;

const formDataEqual = (left: BusinessFormData, right: BusinessFormData): boolean =>
  (Object.keys(INITIAL_FORM_DATA) as Array<keyof BusinessFormData>).every(
    (key) => left[key] === right[key]
  );

const General: React.FC = () => {
  const { t } = useTranslation();
  const { businessType } = useBusinessStore();
  const { user } = useUserStore();
  const { websiteData, contactInfo } = useConfigStore();
  const updateBusiness = useUpdateBusiness();
  const setSlug = useSetBusinessSlug();
  const initialFormData = useMemo<BusinessFormData>(
    () => ({
      ...INITIAL_FORM_DATA,
      storeName: (websiteData?.siteName as string | undefined) || user?.businessName || '',
      contactEmail:
        contactInfo?.contactEmail ||
        (websiteData?.contactEmail as string | undefined) ||
        user?.email ||
        '',
      telephone:
        contactInfo?.contactNumber || (websiteData?.contactNumber as string | undefined) || '',
      companyName: (websiteData?.siteName as string | undefined) || user?.businessName || '',
      address: (websiteData?.address as string | undefined) || '',
      city: (websiteData?.city as string | undefined) || '',
      zipcode: (websiteData?.zipcode as string | undefined) || '',
      mapsUrl: (websiteData?.mapsUrl as string | undefined) || '',
      website: (websiteData?.website as string | undefined) || '',
      country: (websiteData?.country as string | undefined) || 'Lebanon',
      senderEmail:
        (websiteData?.senderEmail as string | undefined) || contactInfo?.contactEmail || '',
    }),
    [user, websiteData, contactInfo]
  );
  const initialSlug =
    typeof (websiteData as { slug?: unknown } | undefined)?.slug === 'string'
      ? (websiteData as { slug: string }).slug
      : '';
  const [formData, setFormData] = useState<BusinessFormData>(INITIAL_FORM_DATA);
  const [slug, setSlugInput] = useState('');

  useEffect(() => {
    setFormData(initialFormData);
  }, [initialFormData]);

  useEffect(() => {
    setSlugInput(initialSlug);
  }, [initialSlug]);

  const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const isFormDirty = !formDataEqual(formData, initialFormData);
  const isSlugDirty = slug !== initialSlug;
  const hasChanges = isFormDirty || isSlugDirty;
  const isSaving = updateBusiness.isPending || setSlug.isPending;

  const handleSave = async (): Promise<void> => {
    if (isFormDirty) {
      const formResult = await Promise.allSettled([
        updateBusiness.mutateAsync({ data: { ...formData, businessType } }),
      ]);
      if (formResult[0]?.status === 'rejected') return;
    }

    if (isSlugDirty && slug) {
      await Promise.allSettled([setSlug.mutateAsync(slug)]);
    }
  };

  const handleCancel = (): void => {
    setFormData(initialFormData);
    setSlugInput(initialSlug);
  };

  const handleSubmit = (e: React.FormEvent): void => {
    e.preventDefault();
    void handleSave();
  };

  return (
    <>
      <HelmetContainer title={t('settings.general.pageTitle')} />
      <div className={cn('space-y-6', hasChanges && 'pb-24')}>
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
              <CardTitle>{t('settings.general.detailsTitle')}</CardTitle>
              <CardDescription>{t('settings.general.detailsDescription')}</CardDescription>
            </CardHeader>
            <CardContent className="pt-0 space-y-4">
              <div className="space-y-1.5">
                <Label>{t('settings.general.storeName')}</Label>
                <Input name="storeName" value={formData.storeName} onChange={handleInputChange} />
              </div>
              <div className="space-y-1.5">
                <Label>{t('settings.general.contactEmail')}</Label>
                <Input
                  type="email"
                  name="contactEmail"
                  value={formData.contactEmail}
                  onChange={handleInputChange}
                />
              </div>
              <div className="space-y-1.5">
                <Label>{t('settings.general.senderEmail')}</Label>
                <Input
                  type="email"
                  name="senderEmail"
                  value={formData.senderEmail}
                  onChange={handleInputChange}
                />
              </div>
              <div className="space-y-1.5">
                <Label>{t('settings.general.publicSlug')}</Label>
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
                    placeholder={t('settings.general.slugPlaceholder')}
                    onChange={(e) =>
                      setSlugInput(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))
                    }
                  />
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
              <CardTitle>{t('settings.general.addressTitle')}</CardTitle>
              <CardDescription>{t('settings.general.addressDescription')}</CardDescription>
            </CardHeader>
            <CardContent className="pt-0 space-y-4">
              {[
                // Translation keys, not English. These labels sat in the data
                // as literals, so they were invisible to a JSX-text sweep and
                // rendered in English for every reader (rtylr#1266).
                { labelKey: 'settings.general.companyName', name: 'companyName' },
                { labelKey: 'settings.general.website', name: 'website' },
                { labelKey: 'settings.general.telephone', name: 'telephone' },
                { labelKey: 'settings.general.mapsUrl', name: 'mapsUrl' },
                { labelKey: 'settings.general.address', name: 'address' },
                { labelKey: 'settings.general.city', name: 'city' },
                { labelKey: 'settings.general.zipcode', name: 'zipcode' },
              ].map((field) => (
                <div key={field.name} className="space-y-1.5">
                  <Label>{t(field.labelKey)}</Label>
                  <Input
                    name={field.name}
                    value={formData[field.name as keyof BusinessFormData]}
                    onChange={handleInputChange}
                  />
                </div>
              ))}
              <div className="space-y-1.5">
                <Label>{t('settings.general.country')}</Label>
                <Input name="country" value={formData.country} onChange={handleInputChange} />
              </div>
            </CardContent>
          </Card>
        </form>

        <AnimatePresence>
          {hasChanges && (
            <motion.div
              key="save-footer"
              initial={{ y: 96, opacity: 0 }}
              animate={{ y: 0, opacity: 1 }}
              exit={{ y: 96, opacity: 0 }}
              transition={{ type: 'spring', stiffness: 380, damping: 30, mass: 0.6 }}
              className="fixed bottom-4 left-1/2 z-40 max-w-[calc(100%-2rem)] -translate-x-1/2"
            >
              <div
                className="flex items-center gap-2 rounded-full py-2 pl-3 pr-2 shadow-lg sm:gap-4 sm:pl-4"
                style={{
                  background: 'var(--dash-card)',
                  border: '1px solid var(--dash-border)',
                  boxShadow: '0 10px 28px -8px rgba(0,0,0,0.18)',
                  backdropFilter: 'blur(8px)',
                }}
              >
                <div className="flex min-w-0 items-center gap-2">
                  <span
                    className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full"
                    style={{ background: 'var(--accent-muted)', color: 'var(--accent-on-muted)' }}
                  >
                    <Check size={13} />
                  </span>
                  <p
                    className="truncate text-xs sm:text-sm"
                    style={{ color: 'var(--dash-text-muted)' }}
                  >
                    {t('common.unsavedChanges')}
                  </p>
                </div>
                <div className="flex shrink-0 items-center gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    className="rounded-full"
                    onClick={handleCancel}
                    disabled={isSaving}
                  >
                    {t('common.cancel')}
                  </Button>
                  <Button
                    type="button"
                    className="rounded-full"
                    onClick={() => void handleSave()}
                    disabled={isSaving || (isSlugDirty && !slug)}
                  >
                    {isSaving ? t('common.saving') : t('common.saveChanges')}
                  </Button>
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </>
  );
};

export default General;
