import { useSelector } from "react-redux";

const AboutUs = () => {
  const config = useSelector((state) => state.config.aboutContent);

  return (
    <div className="about-us-container">
      {config && (
        <div dangerouslySetInnerHTML={{ __html: config?.aboutcontent }} />
      )}
    </div>
  );
};

export default AboutUs;
