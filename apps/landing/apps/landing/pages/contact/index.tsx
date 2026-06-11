import React, { useState } from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
import { useRouter } from 'next/router';
import { MessageCircle, Phone, Mail, ArrowRight, CheckCircle } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
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

const CONTACT_CHANNELS = [
  {
    icon: MessageCircle,
    label: 'Chat with us',
    value: 'WhatsApp +961 3 940 708',
    href: 'https://wa.me/9613940708',
    description: 'Message us directly on WhatsApp.',
  },
  {
    icon: Phone,
    label: 'Call us',
    value: '+961 3 940 708',
    href: 'tel:+9613940708',
    description: 'Mon–Fri, 9 am–5 pm local time.',
  },
  {
    icon: Mail,
    label: 'Email us',
    value: 'info@voxire.com',
    href: 'mailto:info@voxire.com',
    description: 'We reply within 24 hours.',
  },
];

const ContactPage = (): JSX.Element => {
  const router = useRouter();
  const [form, setForm] = useState<ContactForm>({ name: '', email: '', company: '', message: '' });
  const [errors, setErrors] = useState<ContactFormErrors>({});
  const [submitted, setSubmitted] = useState(false);

  const validate = (): boolean => {
    const next: ContactFormErrors = {};
    if (!form.name.trim()) next.name = 'Name is required.';
    if (!form.email.trim()) {
      next.email = 'Email is required.';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
      next.email = 'Please enter a valid email address.';
    }
    if (!form.message.trim()) next.message = 'Message is required.';
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
          title="Contact | rtylr"
          description="Get in touch with the rtylr team. We respond within 24 hours."
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
              Reach out. We read it all.
            </motion.h1>
            <motion.p className={styles.headerSubtitle} variants={blurFadeIn}>
              Product question, sales inquiry, press request, feature ask, every message lands in
              front of a human. We reply within one business day, usually faster.
            </motion.p>
          </motion.section>

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
                <h2 className={styles.infoTitle}>Contact channels</h2>
                <div className={styles.channelsList}>
                  {CONTACT_CHANNELS.map((channel) => {
                    const Icon = channel.icon;
                    return (
                      <a
                        key={channel.label}
                        href={channel.href}
                        className={styles.channelCard}
                        target="_blank"
                        rel="noopener noreferrer"
                        aria-label={channel.label}
                      >
                        <div className={styles.channelIconWrap}>
                          <Icon size={20} strokeWidth={1.75} />
                        </div>
                        <div className={styles.channelContent}>
                          <span className={styles.channelLabel}>{channel.label}</span>
                          <span className={styles.channelValue}>{channel.value}</span>
                          <span className={styles.channelDescription}>{channel.description}</span>
                        </div>
                      </a>
                    );
                  })}
                </div>

                <div className={styles.hoursBlock}>
                  <span className={styles.hoursLabel}>Business hours</span>
                  <span className={styles.hoursValue}>Monday – Friday</span>
                  <span className={styles.hoursValue}>9 am – 5 pm (GMT+3)</span>
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
                    <h3 className={styles.successTitle}>One more step.</h3>
                    <p className={styles.successBody}>
                      We opened a pre-filled draft in your email app — hit send to get it to us and
                      we'll reply within 24 hours. If nothing opened, email us directly at{' '}
                      <a href="mailto:info@voxire.com">info@voxire.com</a>.
                    </p>
                    <button
                      className={styles.successReset}
                      onClick={() => {
                        setSubmitted(false);
                        setForm({ name: '', email: '', company: '', message: '' });
                      }}
                    >
                      Send another message
                    </button>
                  </div>
                ) : (
                  <form onSubmit={handleSubmit} className={styles.form} noValidate>
                    <h2 className={styles.formTitle}>Send us a message</h2>

                    <div className={styles.fieldRow}>
                      <div className={cn(styles.field, errors.name && styles.fieldError)}>
                        <label htmlFor="contact-name" className={styles.fieldLabel}>
                          Name <span className={styles.required}>*</span>
                        </label>
                        <input
                          id="contact-name"
                          name="name"
                          type="text"
                          className={styles.fieldInput}
                          placeholder="Your full name"
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
                          Email <span className={styles.required}>*</span>
                        </label>
                        <input
                          id="contact-email"
                          name="email"
                          type="email"
                          className={styles.fieldInput}
                          placeholder="your@email.com"
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
                        Company <span className={styles.optional}>(optional)</span>
                      </label>
                      <input
                        id="contact-company"
                        name="company"
                        type="text"
                        className={styles.fieldInput}
                        placeholder="Your company name"
                        value={form.company}
                        onChange={handleChange}
                        autoComplete="organization"
                      />
                    </div>

                    <div className={cn(styles.field, errors.message && styles.fieldError)}>
                      <label htmlFor="contact-message" className={styles.fieldLabel}>
                        Message <span className={styles.required}>*</span>
                      </label>
                      <textarea
                        id="contact-message"
                        name="message"
                        className={styles.fieldTextarea}
                        placeholder="Tell us how we can help…"
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
                      Send message
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
            <p className={styles.stripText}>Need help getting started? Book a personalized demo.</p>
            <button className={styles.stripBtn} onClick={handleDemoNavigation}>
              Book a demo
              <ArrowRight size={16} strokeWidth={2} />
            </button>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default ContactPage;
