import React, { useState } from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import { MessageCircle, Phone, Mail, ArrowRight, CheckCircle } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import DirectAnswer from '../../components/DirectAnswer/DirectAnswer';
import { organizationContactSchema } from '../../lib/structured-data';
import {
  blurFadeIn,
  staggerBlur,
  viewportOnce,
  slideInLeft,
  slideInRight,
} from '../../lib/animations';
import styles from './contact.module.scss';

interface ContactForm {
  name: string;
  email: string;
  company: string;
  message: string;
}

interface ContactFormErrors {
  name?: string;
  email?: string;
  message?: string;
}

/** Address routing, in the order a reader is most likely to need it. */
const ROUTING = ['sales', 'support', 'security', 'partners', 'press'] as const;

const CONTACT_CHANNELS = [
  { icon: MessageCircle, id: 'whatsapp', href: 'https://wa.me/9613940708' },
  { icon: Phone, id: 'phone', href: 'tel:+9613940708' },
  { icon: Mail, id: 'email', href: 'mailto:info@voxire.com' },
];

const ContactPage = (): React.JSX.Element => {
  const router = useRouter();
  const { t } = useTranslation();
  const [form, setForm] = useState<ContactForm>({ name: '', email: '', company: '', message: '' });
  const [errors, setErrors] = useState<ContactFormErrors>({});
  const [submitted, setSubmitted] = useState(false);

  const validate = (): boolean => {
    const next: ContactFormErrors = {};
    if (!form.name.trim()) next.name = t('contact.errors.nameRequired');
    if (!form.email.trim()) {
      next.email = t('contact.errors.emailRequired');
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
      next.email = t('contact.errors.emailInvalid');
    }
    if (!form.message.trim()) next.message = t('contact.errors.messageRequired');
    setErrors(next);
    return Object.keys(next).length === 0;
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>): void => {
    const { name, value } = e.target;
    setForm((prev) => ({ ...prev, [name]: value }));
    if (errors[name as keyof ContactFormErrors]) {
      setErrors((prev) => ({ ...prev, [name]: undefined }));
    }
  };

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>): void => {
    e.preventDefault();
    if (!validate()) return;
    const subject = encodeURIComponent(
      `Contact from ${form.name}${form.company ? ` at ${form.company}` : ''}`
    );
    const body = encodeURIComponent(
      `Name: ${form.name}\nEmail: ${form.email}${form.company ? `\nCompany: ${form.company}` : ''}\n\n${form.message}`
    );
    window.location.href = `mailto:info@voxire.com?subject=${subject}&body=${body}`;
    setSubmitted(true);
  };

  const handleDemoNavigation = (): void => {
    void router.push('/demo');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('contact.meta.title')}
          description={t('contact.meta.description')}
          canonical="https://rtylr.com/contact/"
          structuredData={organizationContactSchema()}
        />

        <div className={styles.contactPage}>
          {/* Header */}
          <motion.section
            className={styles.header}
            initial="hidden"
            animate="visible"
            variants={staggerBlur}
          >
            <motion.h1 className={styles.headerTitle} variants={blurFadeIn}>
              {t('contact.hero.title')}
            </motion.h1>
            <motion.p className={styles.headerSubtitle} variants={blurFadeIn}>
              {t('contact.hero.subtitle')}
            </motion.p>
          </motion.section>

          <DirectAnswer label={t('contact.directAnswer.label')}>
            {t('contact.directAnswer.body')}
          </DirectAnswer>

          {/* Two-column layout */}
          <section className={styles.mainSection}>
            <div className={styles.mainContainer}>
              {/* Left: contact info */}
              <motion.div
                className={styles.contactInfo}
                initial="hidden"
                whileInView="visible"
                viewport={viewportOnce}
                variants={slideInLeft}
              >
                <h2 className={styles.infoTitle}>{t('contact.channels.title')}</h2>
                <div className={styles.channelsList}>
                  {CONTACT_CHANNELS.map((channel) => {
                    const Icon = channel.icon;
                    const label = t(`contact.channels.${channel.id}.label`);
                    return (
                      <a
                        key={channel.id}
                        href={channel.href}
                        className={styles.channelCard}
                        target="_blank"
                        rel="noopener noreferrer"
                        aria-label={label}
                      >
                        <div className={styles.channelIconWrap}>
                          <Icon size={20} strokeWidth={1.75} />
                        </div>
                        <div className={styles.channelContent}>
                          <span className={styles.channelLabel}>{label}</span>
                          <span className={styles.channelValue}>
                            {t(`contact.channels.${channel.id}.value`)}
                          </span>
                          <span className={styles.channelDescription}>
                            {t(`contact.channels.${channel.id}.description`)}
                          </span>
                        </div>
                      </a>
                    );
                  })}
                </div>

                <div className={styles.hoursBlock}>
                  <span className={styles.hoursLabel}>{t('contact.hours.label')}</span>
                  <span className={styles.hoursValue}>{t('contact.hours.days')}</span>
                  <span className={styles.hoursValue}>{t('contact.hours.time')}</span>
                </div>

                {/* Which address for what. The channels above are the general
                    ways in; this is the routing table, so a security report
                    does not sit in a sales inbox. */}
                <h2 className={styles.infoTitle}>{t('contact.routing.title')}</h2>
                <div className={styles.channelsList}>
                  {ROUTING.map((id) => (
                    <div key={id} className={styles.channelCard}>
                      <div className={styles.channelContent}>
                        <span className={styles.channelLabel}>
                          {t(`contact.routing.${id}.label`)}
                        </span>
                        {/* The anchor wraps inline content only. An <a> around
                            a block element is dropped wholesale from the
                            markdown twin, which is how the existing channel
                            cards became invisible to agents. */}
                        <a
                          className={styles.channelValue}
                          href={`mailto:${t(`contact.routing.${id}.value`)}`}
                        >
                          {t(`contact.routing.${id}.value`)}
                        </a>
                        <span className={styles.channelDescription}>
                          {t(`contact.routing.${id}.description`)}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </motion.div>

              {/* Right: form */}
              <motion.div
                className={styles.formSection}
                initial="hidden"
                whileInView="visible"
                viewport={viewportOnce}
                variants={slideInRight}
              >
                {submitted ? (
                  <div className={styles.successState}>
                    <CheckCircle size={48} strokeWidth={1.5} className={styles.successIcon} />
                    <h3 className={styles.successTitle}>{t('contact.success.title')}</h3>
                    <p className={styles.successBody}>
                      {t('contact.success.body')}{' '}
                      <a href="mailto:info@voxire.com">info@voxire.com</a>.
                    </p>
                    <button
                      className={styles.successReset}
                      onClick={() => {
                        setSubmitted(false);
                        setForm({ name: '', email: '', company: '', message: '' });
                      }}
                    >
                      {t('contact.success.reset')}
                    </button>
                  </div>
                ) : (
                  <form onSubmit={handleSubmit} className={styles.form} noValidate>
                    <h2 className={styles.formTitle}>{t('contact.form.title')}</h2>

                    <div className={styles.fieldRow}>
                      <div className={cn(styles.field, errors.name && styles.fieldError)}>
                        <label htmlFor="contact-name" className={styles.fieldLabel}>
                          {t('contact.form.name.label')}{' '}
                          <span className={styles.required}>{t('contact.form.required')}</span>
                        </label>
                        <input
                          id="contact-name"
                          name="name"
                          type="text"
                          className={styles.fieldInput}
                          placeholder={t('contact.form.name.placeholder')}
                          value={form.name}
                          onChange={handleChange}
                          autoComplete="name"
                          aria-invalid={!!errors.name}
                          aria-describedby={errors.name ? 'contact-name-error' : undefined}
                        />
                        {errors.name && (
                          <span id="contact-name-error" className={styles.errorMsg}>
                            {errors.name}
                          </span>
                        )}
                      </div>

                      <div className={cn(styles.field, errors.email && styles.fieldError)}>
                        <label htmlFor="contact-email" className={styles.fieldLabel}>
                          {t('contact.form.email.label')}{' '}
                          <span className={styles.required}>{t('contact.form.required')}</span>
                        </label>
                        <input
                          id="contact-email"
                          name="email"
                          type="email"
                          className={styles.fieldInput}
                          placeholder={t('contact.form.email.placeholder')}
                          value={form.email}
                          onChange={handleChange}
                          autoComplete="email"
                          aria-invalid={!!errors.email}
                          aria-describedby={errors.email ? 'contact-email-error' : undefined}
                        />
                        {errors.email && (
                          <span id="contact-email-error" className={styles.errorMsg}>
                            {errors.email}
                          </span>
                        )}
                      </div>
                    </div>

                    <div className={styles.field}>
                      <label htmlFor="contact-company" className={styles.fieldLabel}>
                        {t('contact.form.company.label')}{' '}
                        <span className={styles.optional}>{t('contact.form.optional')}</span>
                      </label>
                      <input
                        id="contact-company"
                        name="company"
                        type="text"
                        className={styles.fieldInput}
                        placeholder={t('contact.form.company.placeholder')}
                        value={form.company}
                        onChange={handleChange}
                        autoComplete="organization"
                      />
                    </div>

                    <div className={cn(styles.field, errors.message && styles.fieldError)}>
                      <label htmlFor="contact-message" className={styles.fieldLabel}>
                        {t('contact.form.message.label')}{' '}
                        <span className={styles.required}>{t('contact.form.required')}</span>
                      </label>
                      <textarea
                        id="contact-message"
                        name="message"
                        className={styles.fieldTextarea}
                        placeholder={t('contact.form.message.placeholder')}
                        rows={5}
                        value={form.message}
                        onChange={handleChange}
                        aria-invalid={!!errors.message}
                        aria-describedby={errors.message ? 'contact-message-error' : undefined}
                      />
                      {errors.message && (
                        <span id="contact-message-error" className={styles.errorMsg}>
                          {errors.message}
                        </span>
                      )}
                    </div>

                    <button type="submit" className={styles.submitBtn}>
                      {t('contact.form.submit')}
                    </button>
                  </form>
                )}
              </motion.div>
            </div>
          </section>

          {/* Bottom strip */}
          <motion.section
            className={styles.bottomStrip}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={blurFadeIn}
          >
            <p className={styles.stripText}>{t('contact.strip.text')}</p>
            <button className={styles.stripBtn} onClick={handleDemoNavigation}>
              {t('common.cta.bookDemo')}
              <ArrowRight size={16} strokeWidth={2} />
            </button>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default ContactPage;
