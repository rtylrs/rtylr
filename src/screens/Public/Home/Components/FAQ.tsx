import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useSelector } from "react-redux";
import VbitezButton from "../../../../components/Buttons/Buttons";
import { useNavigate } from "react-router-dom";
import { RootState } from "../../../../redux/store";

function Faq() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [activeIndex, setActiveIndex] = useState(null);
  const faqs = useSelector((state: RootState) => state.config.faqs);

  const toggleAnswer = (index) => {
    setActiveIndex(activeIndex === index ? null : index);
  };

  return (
    <div className="faq-section">
      <h2 className="vbitez-title">{t("QandA.title")}</h2>
      <div className="faq-container">
        {faqs &&
          faqs.slice(0, 8).map((item: any, index: number) => (
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
      <div
        style={{
          marginTop: "40px",
        }}
      >
        <VbitezButton variant="default" onClick={() => navigate("/faqs")}>
          View All
        </VbitezButton>
      </div>
    </div>
  );
}
export default Faq;
