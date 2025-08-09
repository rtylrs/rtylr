import { useSelector } from 'react-redux';
import { RootState } from '../redux/store';

const AboutUs = (): JSX.Element => {
  const config = useSelector((state: RootState) => state.config.aboutContent);

  return (
    <div className="about-us-container">
      {config && <div dangerouslySetInnerHTML={{ __html: config }} />}
    </div>
  );
};

export default AboutUs;
