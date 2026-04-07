import { Button } from '@voxire/components';
import { useState } from "react";
import ReactQuill from "react-quill";
import "react-quill/dist/quill.snow.css"; // Import the styles
import { useDispatch, useSelector } from "react-redux";
import { updateAboutContent } from "../../../redux/reducers/configSlice";
import notify from "../../../lib/notify/notify";
import { AppDispatch, RootState } from "../../../redux/store";

function AboutConfig() {
  const config = useSelector((state: RootState) => state.config.aboutContent);
  const [text, setText] = useState(config);
  const [isTextModified, setIsTextModified] = useState(false);
  const dispatch = useDispatch<AppDispatch>();

  const handleChange = (value) => {
    setText(value);
    setIsTextModified(value !== config?.aboutcontent);
  };

  const handleSubmit = (event) => {
    event.preventDefault();
    dispatch(
      updateAboutContent({
        aboutContent: text,
      })
    );
    setIsTextModified(false);
    notify.success("Text updated successfully");
  };

  return (
    <div className="about-config">
      <span>
        <h1>About Configuration</h1>
        <form onSubmit={handleSubmit}>
          <ReactQuill value={text} onChange={handleChange} />
          {isTextModified && (
            <Button
              variant="contained"
              type="submit"
              color="primary"
              style={{
                marginTop: 20,
              }}
            >
              Update Text
            </Button>
          )}
        </form>
      </span>
    </div>
  );
}

export default AboutConfig;
