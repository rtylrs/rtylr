import { useTranslation } from "react-i18next";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faRocketchat } from "@fortawesome/free-brands-svg-icons";
import { faPhone } from "@fortawesome/free-solid-svg-icons";
import { faEnvelope } from "@fortawesome/free-solid-svg-icons";
import React from "react";

const Contact = (): JSX.Element => {
  const { t } = useTranslation();

  return (
    <div className="ContactUs">
      <div className="contactus-container">
        <div className="contactus-card">
          <div className="contactus-icon">
            <FontAwesomeIcon icon={faRocketchat} />
          </div>
          <h2>
            <b>Chat with us</b>
          </h2>
          <p>
            We're waiting to help you every Monday-Friday from 9 am to 5 pm EST
            easily.
          </p>
        </div>
        <div className="contactus-card">
          <div className="contactus-icon">
            <FontAwesomeIcon icon={faPhone} />
          </div>
          <h2>
            <b>Give us a call</b>
          </h2>

          <p>
            Give us a ring at (+961 71 642 565). Every monday-friday from 9 am
            to 5 pm.
          </p>
        </div>
        <div className="contactus-card">
          <div className="contactus-icon">
            <FontAwesomeIcon icon={faEnvelope} />
          </div>
          <h2>
            <b>Email Us</b>
          </h2>

          <p>
            Drop us an email at info@voxire.com and you'll receive a reply
            within 24 hours.
          </p>
        </div>
      </div>
      <div className="contact-form-container">
        <div className="contact-form-text">
          <h1>Fill out this form, We‘ll quickly get back to you</h1>
          <p>
            We are here to help you! Tell us how we can help and we’ll get in
            touch within the next 24hrs with an expert.
          </p>
          <div className="testimonial">
            <div className="stars">
              <span>★★★★★</span>
            </div>
            <p>
              "Rtylr transformed our restaurant experience with their digital
              menu system. Our customers love it and we see increased efficiency
              in our service."
            </p>
            <div className="testimonial-author">
              <img src="/images/Al-Abed_Mohammad2.jpg" alt="Author" />
              <div className="author-details">
                <h3>Aalawi Habeeb Galbi</h3>
                <p>CEO & Co-founder @ Tiktok</p>
              </div>
            </div>
          </div>
        </div>
        <div className="contact-form">
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
                  resize: "none",
                }}
              ></textarea>
            </label>
            <button type="submit">Submit</button>
          </form>
        </div>
      </div>
    </div>
  );
};
export default Contact;
