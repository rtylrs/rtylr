import { faHourglassHalf } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

const Addons = () => {
  return (
    <div className="addons">
      <FontAwesomeIcon icon={faHourglassHalf} />
      <h1>Coming Soon</h1>
      <p>
        We are excited to announce that Add-Ons will be available soon. Please
        stay tuned for updates!
      </p>
    </div>
  );
};

export default Addons;
