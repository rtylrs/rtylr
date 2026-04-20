import React, { useState, useMemo } from 'react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { Input, Select, Button } from '@voxire/components';
import countries from 'i18n-iso-countries';
import enLocale from 'i18n-iso-countries/langs/en.json';
import styles from './demo.module.scss';

// Register English locale for country names
countries.registerLocale(enLocale);

interface FormData {
  email: string;
  country: string;
  reason: string;
  companySize: string;
}

const DemoPage = (): JSX.Element => {
  const [formData, setFormData] = useState<FormData>({
    email: '',
    country: '',
    reason: '',
    companySize: '',
  });

  const [errors, setErrors] = useState<Partial<FormData>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

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

  const reasonOptions = [
    {
      id: 'enterprise',
      title: 'Explore Rtylr Enterprise',
      value: 'Explore Rtylr Enterprise',
    },
    {
      id: 'usecase',
      title: 'Explore a Rtylr Use Case',
      value: 'Explore a Rtylr Use Case',
    },
    { id: 'partnerships', title: 'Partnerships', value: 'Partnerships' },
    {
      id: 'services',
      title: 'Professional Services',
      value: 'Professional Services',
    },
  ];

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
      newErrors.email = 'Work email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = 'Please enter a valid email address';
    }

    if (!formData.country) {
      newErrors.country = 'Please select your location';
    }

    if (!formData.reason) {
      newErrors.reason = 'Please select a reason for demo';
    }

    if (!formData.companySize) {
      newErrors.companySize = 'Please select company size';
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

    // Simulate API call
    setTimeout(() => {
      console.log('Form submitted:', formData);
      alert('Thank you! We will contact you shortly to schedule your demo.');
      setIsSubmitting(false);
      // Reset form
      setFormData({
        email: '',
        country: '',
        reason: '',
        companySize: '',
      });
    }, 1000);
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="Book a demo | rtylr"
          description="See rtylr running against real data. A live 30-minute walkthrough with someone who actually built it."
          canonical="https://rtylr.com/demo/"
        />
        <div className={styles.demoPage}>
          <div className={styles.container}>
            <div className={styles.formWrapper}>
              <h1 className={styles.title}>See rtylr on your data.</h1>
              <p className={styles.subtitle}>
                A 30-minute walkthrough with someone who actually built it. Bring your menu, your
                SKU list, your PDF of invoices, we&apos;ll show you what the switch would look like.
              </p>

              <form onSubmit={handleSubmit} className={styles.form}>
                <div className={styles.formGroup}>
                  <Input
                    type="email"
                    label="WORK EMAIL"
                    placeholder="Your work email"
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
                    label="WHERE ARE YOU LOCATED?"
                    placeholder="Select your country"
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
                    label="REASON FOR DEMO"
                    placeholder="Select reason"
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
                    label="HOW MANY EMPLOYEES WORK AT YOUR COMPANY?"
                    placeholder="Select company size"
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
              </form>
            </div>
          </div>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default DemoPage;
