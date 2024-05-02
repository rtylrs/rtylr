import { useSelector } from "react-redux";

const AboutUs = () => {
  const config = useSelector((state) => state.config.aboutContent);

  console.log(config);

  return (
    <div className="about-us-container">
      {config && <div dangerouslySetInnerHTML={{ __html: config }} />}
    </div>
  );
};

export default AboutUs;
