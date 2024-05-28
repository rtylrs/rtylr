import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faChartPie,
  faCoffee,
  faCog,
  faCreditCard,
  faMagicWandSparkles,
  faPhoneVolume,
  faPlus,
  faQrcode,
  faUtensils,
} from "@fortawesome/free-solid-svg-icons"; // Example icon
import PropTypes from "prop-types";
import { Link } from "react-router-dom";

function QuickLinks() {
  return (
    <div className="quick-links">
      <QuickLink icon={faUtensils} title="Menu" to="/app/restaurant/menu" />
      <QuickLink
        icon={faChartPie}
        title="Reports"
        to="/app/restaurant/reports"
      />
      <QuickLink
        icon={faQrcode}
        title="Qr Codes"
        to="/app/restaurant/qr-codes/"
      />
      <QuickLink
        icon={faPlus}
        title="Create Qr"
        to="/app/restaurant/qr-codes/create"
      />
      <QuickLink
        icon={faMagicWandSparkles}
        title="Design"
        to="/app/restaurant/design"
      />
      <QuickLink
        icon={faCreditCard}
        title="Billing"
        to="/app/restaurant/billing"
      />
      <QuickLink icon={faCog} title="Settings" to="/app/restaurant/settings" />
      <QuickLink
        icon={faPhoneVolume}
        title="Support"
        to="/app/restaurant/settings"
      />
    </div>
  );
}

const QuickLink = ({ icon, title, to }) => (
  <Link className="quick-link" to={to}>
    <FontAwesomeIcon icon={icon} />
    <span>{title}</span>
  </Link>
);

QuickLink.propTypes = {
  icon: PropTypes.object.isRequired,
  title: PropTypes.string.isRequired,
  to: PropTypes.string.isRequired,
};

export default QuickLinks;
