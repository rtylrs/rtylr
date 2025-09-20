import { useConfigStore } from '../stores/configStore';

const AboutUs = (): JSX.Element => {
  const aboutContent = useConfigStore((state) => state.aboutContent);

  return (
    <div className="about-us-container">
      {aboutContent && <div dangerouslySetInnerHTML={{ __html: aboutContent }} />}
    </div>
  );
};

export default AboutUs;
