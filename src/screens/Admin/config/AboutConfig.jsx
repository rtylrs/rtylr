import { useState } from "react";
import ReactQuill from "react-quill";
import "react-quill/dist/quill.snow.css"; // Import the styles
import { useDispatch, useSelector } from "react-redux";
import { UpdateAboutContent } from "../../../redux/reducers/configSlice";

function AboutConfig() {
  const config = useSelector((state) => state.config.aboutContent);
  const [text, setText] = useState(config?.aboutcontent);
  const [isTextModified, setIsTextModified] = useState(false);
  const dispatch = useDispatch();

  const handleChange = (value) => {
    setText(value);
    setIsTextModified(value !== config?.aboutcontent);
  };

  const handleSubmit = (event) => {
    event.preventDefault();
    dispatch(
      UpdateAboutContent({
        aboutcontent: text,
      })
    );
    setIsTextModified(false);
  };

  return (
    <div className="about-config">
      <div className="InsuficientPackage">
        <p>Your current package does not support this feature.</p>
      </div>
      <span>
        <div className="overlay-about" style={{ display: "block" }}></div>
        <h1>About Configuration</h1>
        <form onSubmit={handleSubmit}>
          <ReactQuill value={text} onChange={handleChange} />
          {isTextModified && <button type="submit">Update Text</button>}
        </form>
      </span>
    </div>
  );
}

export default AboutConfig;
