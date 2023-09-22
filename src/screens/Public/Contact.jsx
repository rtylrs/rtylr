import { useState } from "react";
import { motion } from "framer-motion";
import { db } from "../../firebase";

const Contact = () => {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [message, setMessage] = useState("");
  const [submitted, setSubmitted] = useState(false);

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: { opacity: 1, transition: { delay: 0.2, duration: 0.8 } },
    exit: { opacity: 0, transition: { ease: "easeInOut" } },
  };

  const itemVariants = {
    hidden: { y: -20, opacity: 0 },
    visible: {
      y: 0,
      opacity: 1,
      transition: { duration: 0.3, ease: "easeInOut" },
    },
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!name || !email || !message) {
      return;
    }

    try {
      await db.collection("contacts").add({
        name,
        email,
        phone,
        message,
      });

      setSubmitted(true);
    } catch (error) {
      console.error("Error adding contact:", error);
    }
  };

  return (
    <motion.div
      className="contact-container"
      variants={containerVariants}
      initial="hidden"
      animate="visible"
      exit="exit"
    >
      <motion.h2 variants={itemVariants}>Contact Us</motion.h2>
      <motion.p variants={itemVariants}>
        Feel free to get in touch with us!
      </motion.p>
      <div className="contact-content">
        <motion.form
          className="contact-form"
          variants={itemVariants}
          onSubmit={handleSubmit}
        >
          <input
            type="text"
            placeholder="Name"
            className="input-field"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          <input
            type="email"
            placeholder="Email"
            className="input-field"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          <input
            type="tel"
            placeholder="Phone Number (Include Country Code)"
            className="input-field"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
          />
          <textarea
            placeholder="Message"
            rows="4"
            className="input-field"
            value={message}
            onChange={(e) => setMessage(e.target.value)}
          ></textarea>
          <button
            type="submit"
            className="main__btn"
            disabled={submitted}
            style={{
              display: "flex",
              justifyContent: "center",
              alignItems: "center",
              width: "60%",
              margin: "0 auto",
            }}
          >
            {submitted ? "Submitted" : "Submit"}
          </button>
        </motion.form>
      </div>
    </motion.div>
  );
};

export default Contact;
