import { useState } from "react";
import { useTranslation } from "react-i18next";

function Faq() {
  const {t} = useTranslation()
  const [activeIndex, setActiveIndex] = useState(null);

  const faqData = [
    {
      question: "What is VBitez?",
      answer:
        "VBitez is an innovative digital menu solution for restaurants, combining the best features of Restbits and Omega to enhance customer experience and streamline restaurant operations.",
    },
    {
      question: "What makes VBitez different from Restbits and Omega?",
      answer:
        "VBitez not only integrates the best features of both Restbits and Omega but also addresses their gaps. We offer advanced customization, seamless POS integration, and unique features designed to improve both user experience and restaurant workflows.",
    },
    {
      question: "How does VBitez improve customer experience?",
      answer:
        "VBitez provides a visually appealing digital menu with images and descriptions, easy QR code access, and real-time updates, ensuring customers have the latest information and a delightful ordering experience.",
    },
    {
      question: "Can VBitez handle online ordering and menu management?",
      answer:
        "Yes, VBitez offers comprehensive online ordering functionality along with easy menu management. Restaurants can quickly update their menus and manage orders efficiently.",
    },
    {
      question: "What unique features does VBitez offer?",
      answer:
        "VBitez introduces features such as detailed descriptions of menu items, including ingredients and allergens, filtering or searching the menu by category (e.g., vegetarian, gluten-free), menu translation into different languages, order customization options, bill splitting, feedback and review options, pre-ordering food, and integration with loyalty programs or rewards systems.",
    },
  ];

  const toggleAnswer = (index) => {
    setActiveIndex(activeIndex === index ? null : index);
  };

  return (
    <div className="faq-section">
      <h2>{t("QandA.title")}</h2>
      <div className="faq-container">
        {faqData.map((item, index) => (
          <div key={index} className="faq-item">
            <div className="faq-question" onClick={() => toggleAnswer(index)}>
              <span>Q. {item.question}</span>
              <span className="faq-icon">
                {activeIndex === index ? "−" : "+"}
              </span>
            </div>
            {activeIndex === index && (
              <div className="faq-answer">
                <p>{item.answer}</p>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
export default Faq;
