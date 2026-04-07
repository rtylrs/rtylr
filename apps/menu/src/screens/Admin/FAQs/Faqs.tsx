import { useState } from "react";
import { useSelector } from "react-redux";
import { Button } from '@voxire/components';
import { useNavigate } from "react-router-dom";
import { RootState } from "../../../redux/store";

function Faqs() {
  const [activeIndex, setActiveIndex] = useState(null);
  const faqs = useSelector((state: RootState) => state.config.faqs);
  const navigate = useNavigate();

  const toggleAnswer = (index: any) => {
    setActiveIndex(activeIndex === index ? null : index);
  };

  return (
    <div className="faq-section">
      <div
        style={{
          display: "flex",
          justifyContent: "flex-end",
          marginBottom: "20px",
        }}
      >
        <Button
          variant="default"
          onClick={() => {
            navigate("/manage/sections/faqs/add");
          }}
        >
          Add FAQ
        </Button>
      </div>
      <div className="faq-container">
        {faqs &&
          faqs.slice(0, 8).map((item, index) => (
            <div key={index} className="faq-item">
              <div className="faq-question" onClick={() => toggleAnswer(index)}>
                <span>Q. {item.question}</span>
              </div>
              <div className="faq-answer">
                <p>A. {item.answer}</p>
              </div>
            </div>
          ))}
      </div>
    </div>
  );
}
export default Faqs;
