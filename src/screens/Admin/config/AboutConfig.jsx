import { useState } from "react";
import ReactQuill from "react-quill";
import "react-quill/dist/quill.snow.css"; // Import the styles
import { useDispatch, useSelector } from "react-redux";
import { updateAboutContent } from "../../../redux/reducers/configSlice";
import { Button } from "@mui/material";
import notify from "../../../lib/notify/notify";

function AboutConfig() {
  const config = useSelector((state) => state.config.aboutContent);
  const [text, setText] = useState(config);
  const [isTextModified, setIsTextModified] = useState(false);
  const dispatch = useDispatch();

  const handleChange = (value) => {
    setText(value);
    setIsTextModified(value !== config?.aboutcontent);
  };

  const handleSubmit = (event) => {
    event.preventDefault();
    // dispatch(
    //   updateAboutContent({
    //     aboutContent: text,
    //   })
    // );
    // setIsTextModified(false);
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
