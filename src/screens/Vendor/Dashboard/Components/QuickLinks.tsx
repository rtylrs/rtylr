import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faChartPie,
  faCoffee,
  faCog,
  faCreditCard,
  faFolder,
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
      <QuickLink icon={faUtensils} title="Menu" to="/app/business/menu" />
      <QuickLink icon={faFolder} title="My Menus" to="/app/business/menus" />
      <QuickLink icon={faChartPie} title="Reports" to="/app/business/reports" />
      <QuickLink
        icon={faQrcode}
        title="Qr Codes"
        to="/app/business/qr-codes/"
      />
      <QuickLink
        icon={faPlus}
        title="Create Qr"
        to="/app/business/qr-codes/create"
      />
      <QuickLink
        icon={faMagicWandSparkles}
        title="Design"
        to="/app/business/design"
      />
      <QuickLink
        icon={faCreditCard}
        title="Billing"
        to="/app/business/billing"
      />
      <QuickLink icon={faCog} title="Settings" to="/app/business/settings" />
      <QuickLink icon={faPhoneVolume} title="Support" to="/contact" />
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
