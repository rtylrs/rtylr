import React, { useState, useMemo } from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { Input, Select, Button } from '@voxire/components';
import countries from 'i18n-iso-countries';
import enLocale from 'i18n-iso-countries/langs/en.json';
import { contactPageSchema } from '../../lib/structured-data';
import { blurFadeIn, staggerBlur } from '../../lib/animations';
import { SplitWords } from '../../components/motion/SplitWords';
import styles from './demo.module.scss';

// Register English locale for country names
countries.registerLocale(enLocale);

interface FormData {
  email: string;
  country: string;
  reason: string;
  companySize: string;
}

const DemoPage = (): React.JSX.Element => {
  const { t } = useTranslation();
  const [formData, setFormData] = useState<FormData>({
    email: '',
    country: '',
    reason: '',
    companySize: '',
  });

  const [errors, setErrors] = useState<Partial<FormData>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  // Rendered as an inline success block instead of alert() so the
  // form doesn't yank the user's attention with a native dialog.
  const [submitted, setSubmitted] = useState(false);

  // Generate country options from i18n-iso-countries
  const countryOptions = useMemo(() => {
    const countryObj = countries.getNames('en', { select: 'official' });
    return Object.entries(countryObj)
      .map(([code, name]) => ({
        id: code.toLowerCase(),
        title: name,
        value: name,
      }))
      .sort((a, b) => a.title.localeCompare(b.title));
  }, []);

  // The id and value are submitted, so they stay in English: they are
  // data. Only the visible title is translated.
  const reasonTitles = t('demo.reasons', { returnObjects: true }) as unknown as string[];
  const reasonOptions = ['enterprise', 'usecase', 'partnerships', 'services'].map((id, i) => ({
    id,
    title: reasonTitles[i],
    value: id,
  }));


  const companySizeOptions = [
    { id: '1-25', title: '1-25', value: '1-25' },
    { id: '26-50', title: '26-50', value: '26-50' },
    { id: '51-100', title: '51-100', value: '51-100' },
    { id: '101-300', title: '101-300', value: '101-300' },
    { id: '301-500', title: '301-500', value: '301-500' },
    { id: '501-1000', title: '501-1,000', value: '501-1,000' },
    { id: '1001-2500', title: '1,001-2,500', value: '1,001-2,500' },
    { id: '2501-5000', title: '2,501-5,000', value: '2,501-5,000' },
    { id: '5001+', title: '5,001+', value: '5,001+' },
  ];

  const validateForm = (): boolean => {
    const newErrors: Partial<FormData> = {};

    if (!formData.email) {
      newErrors.email = t('demo.errors.emailRequired');
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = t('demo.errors.emailInvalid');
    }

    if (!formData.country) {
      newErrors.country = t('demo.errors.country');
    }

    if (!formData.reason) {
      newErrors.reason = t('demo.errors.reason');
    }

    if (!formData.companySize) {
      newErrors.companySize = t('demo.errors.size');
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setIsSubmitting(true);
    const subject = encodeURIComponent('Demo request from rtylr.com');
    const body = encodeURIComponent(
      `Email: ${formData.email}\nCountry: ${formData.country}\nCompany size: ${formData.companySize}\nReason: ${formData.reason}`
    );
    window.location.href = `mailto:sales@rtylr.com?subject=${subject}&body=${body}`;
    setSubmitted(true);
    setFormData({ email: '', country: '', reason: '', companySize: '' });
    setIsSubmitting(false);
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('demo.meta.title')}
          description={t('demo.meta.description')}
          canonical="https://rtylr.com/demo/"
          structuredData={contactPageSchema({
            name: t('demo.meta.schemaName'),
            path: '/demo/',
            description: t('demo.meta.description'),
          })}
        />
        <div className={styles.demoPage}>
          <div className={styles.container}>
            <motion.div
              className={styles.formWrapper}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              {submitted ? (
                // Inline success state - nicer than a native alert().
                // Keeps the page branded and tells the user the exact
                // next step so they're not left wondering.
                <>
                  <SplitWords as="h1" className={styles.title} text={t('demo.success.title')} delay={0.05} />
                  <p className={styles.subtitle}>
                    {t('demo.success.body')}{' '}
                    <strong>
                      <a href="mailto:sales@rtylr.com">sales@rtylr.com</a>
                    </strong>
                    .
                  </p>
                </>
              ) : (
                <>
                  <SplitWords as="h1" className={styles.title} text={t('demo.hero.title')} delay={0.05} />
                  <motion.p className={styles.subtitle} variants={blurFadeIn}>
                    {t('demo.hero.subtitle')}
                  </motion.p>

                  <motion.form onSubmit={handleSubmit} className={styles.form} variants={blurFadeIn}>
                    <div className={styles.formGroup}>
                      <Input
                        type="email"
                        label={t('demo.form.emailLabel')}
                        placeholder={t('demo.form.emailPlaceholder')}
                        value={formData.email}
                        onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                        error={!!errors.email}
                        helperText={errors.email}
                        required
                        className={styles.input}
                      />
                    </div>

                    <div className={styles.formGroup}>
                      <Select
                        label={t('demo.form.countryLabel')}
                        placeholder={t('demo.form.countryPlaceholder')}
                        options={countryOptions}
                        value={formData.country}
                        onChange={(value) => setFormData({ ...formData, country: value })}
                        error={!!errors.country}
                        helperText={errors.country}
                        required
                        className={styles.select}
                      />
                    </div>

                    <div className={styles.formGroup}>
                      <Select
                        label={t('demo.form.reasonLabel')}
                        placeholder={t('demo.form.reasonPlaceholder')}
                        options={reasonOptions}
                        value={formData.reason}
                        onChange={(value) => setFormData({ ...formData, reason: value })}
                        error={!!errors.reason}
                        helperText={errors.reason}
                        required
                        className={styles.select}
                      />
                    </div>

                    <div className={styles.formGroup}>
                      <Select
                        label={t('demo.form.sizeLabel')}
                        placeholder={t('demo.form.sizePlaceholder')}
                        options={companySizeOptions}
                        value={formData.companySize}
                        onChange={(value) => setFormData({ ...formData, companySize: value })}
                        error={!!errors.companySize}
                        helperText={errors.companySize}
                        required
                        className={styles.select}
                      />
                    </div>

                    <p className={styles.privacyText}>
                      By submitting this form, you agree to our{' '}
                      <a href="/legal/privacy-policy" target="_blank" rel="noopener noreferrer">
                        privacy policy
                      </a>{' '}
                      and consent to receiving marketing communications from Rtylr.
                    </p>

                    <Button type="submit" variant="soft" disabled={isSubmitting}>
                      {isSubmitting ? 'SUBMITTING...' : 'BOOK A DEMO'}
                    </Button>
                  </motion.form>
                </>
              )}
            </motion.div>
          </div>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default DemoPage;
