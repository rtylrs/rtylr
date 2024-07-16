import { useState } from "react";
import useAsync from "../../../hooks/useAsync";
import adminApiExport from "../../../api/admin";
import { useNavigate } from "react-router-dom";
import { useDispatch } from "react-redux";
import { fetchFaqs } from "../../../redux/reducers/configSlice";
import { AppDispatch } from "../../../redux/store";

const AddFaq = () => {
  const [question, setQuestion] = useState({ en: "", ar: "", fr: "" });
  const [answer, setAnswer] = useState({ en: "", ar: "", fr: "" });
  const [videoUrl, setVideoUrl] = useState("");
  const navigate = useNavigate();
  const dispatch = useDispatch<AppDispatch>();

  const addFaqFunc = useAsync({
    fn: adminApiExport.addFaq,
    onSuccess: () => {
      alert("FAQ added successfully");
      dispatch(fetchFaqs());
      navigate("/manage/sections/faqs");
    },
    onError: () => {
      alert(addFaqFunc.error);
    },
    callOnMount: false,
  });

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const faqData = { question, answer, videoUrl };
      await addFaqFunc.main(faqData);
    } catch (error) {
      console.error("Error adding FAQ:", error);
    }
  };

  const handleQuestionChange = (lang, value) => {
    setQuestion({ ...question, [lang]: value });
  };

  const handleAnswerChange = (lang, value) => {
    setAnswer({ ...answer, [lang]: value });
  };

  return (
    <div className="add-faq-container">
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label className="form-label">Question (English):</label>
          <input
            type="text"
            className="form-input"
            value={question.en}
            onChange={(e) => handleQuestionChange("en", e.target.value)}
            required
          />
        </div>
        <div className="form-group">
          <label className="form-label">Question (Arabic):</label>
          <input
            type="text"
            className="form-input"
            value={question.ar}
            onChange={(e) => handleQuestionChange("ar", e.target.value)}
          />
        </div>
        <div className="form-group">
          <label className="form-label">Question (French):</label>
          <input
            type="text"
            className="form-input"
            value={question.fr}
            onChange={(e) => handleQuestionChange("fr", e.target.value)}
          />
        </div>
        <div className="form-group">
          <label className="form-label">Answer (English):</label>
          <textarea
            className="form-textarea"
            value={answer.en}
            onChange={(e) => handleAnswerChange("en", e.target.value)}
            required
          />
        </div>
        <div className="form-group">
          <label className="form-label">Answer (Arabic):</label>
          <textarea
            className="form-textarea"
            value={answer.ar}
            onChange={(e) => handleAnswerChange("ar", e.target.value)}
          />
        </div>
        <div className="form-group">
          <label className="form-label">Answer (French):</label>
          <textarea
            className="form-textarea"
            value={answer.fr}
            onChange={(e) => handleAnswerChange("fr", e.target.value)}
          />
        </div>
        <div className="form-group">
          <label className="form-label">Video URL:</label>
          <input
            type="url"
            className="form-input"
            value={videoUrl}
            onChange={(e) => setVideoUrl(e.target.value)}
          />
        </div>
        <button type="submit" className="form-button">
          Add FAQ
        </button>
      </form>
    </div>
  );
};

export default AddFaq;
