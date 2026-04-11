import React from 'react';
import { HiChat, HiPhone, HiMail } from 'react-icons/hi';
import { useTranslation } from 'react-i18next';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import styles from './contact.module.scss';

const ContactPage = (): JSX.Element => {
  const { t } = useTranslation();

  return (
    <AppLayout>
      <MainLayout>
        <div className={styles.ContactUs}>
          <div className={styles['contactus-container']}>
            <div className={styles['contactus-card']}>
              <div className={styles['contactus-icon']}>
                <HiChat />
              </div>
              <h2>
                <b>Chat with us</b>
              </h2>
              <p>We're waiting to help you every Monday-Friday from 9 am to 5 pm EST easily.</p>
            </div>
            <div className={styles['contactus-card']}>
              <div className={styles['contactus-icon']}>
                <HiPhone />
              </div>
              <h2>
                <b>Give us a call</b>
              </h2>
              <p>Give us a ring at (+961 71 642 565). Every monday-friday from 9 am to 5 pm.</p>
            </div>
            <div className={styles['contactus-card']}>
              <div className={styles['contactus-icon']}>
                <HiMail />
              </div>
              <h2>
                <b>Email Us</b>
              </h2>
              <p>Drop us an email at info@voxire.com and you'll receive a reply within 24 hours.</p>
            </div>
          </div>
          <div className={styles['contact-form-container']}>
            <div className={styles['contact-form-text']}>
              <h1>Fill out this form, We'll quickly get back to you</h1>
              <p>
                We are here to help you! Tell us how we can help and we'll get in touch within the
                next 24hrs with an expert.
              </p>
              <div className={styles.testimonial}>
                <div className={styles.stars}>
                  <span>★★★★★</span>
                </div>
                <p>
                  "Rtylr transformed our restaurant experience with their digital menu system. Our
                  customers love it and we see increased efficiency in our service."
                </p>
                <div className={styles['testimonial-author']}>
                  <img src="/images/Al-Abed_Mohammad2.jpg" alt="Author" />
                  <div className={styles['author-details']}>
                    <h3>Aalawi Habeeb Galbi</h3>
                    <p>CEO & Co-founder @ Tiktok</p>
                  </div>
                </div>
              </div>
            </div>
            <div className={styles['contact-form']}>
              <form>
                <label>
                  Your name
                  <input type="text" placeholder="Enter your full name" />
                </label>
                <label>
                  Email address
                  <input type="email" placeholder="Enter your email" />
                </label>
                <label>
                  Write your message
                  <textarea
                    placeholder="Write us your question here..."
                    style={{
                      resize: 'none',
                    }}
                  ></textarea>
                </label>
                <button type="submit">Submit</button>
              </form>
            </div>
          </div>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default ContactPage;
