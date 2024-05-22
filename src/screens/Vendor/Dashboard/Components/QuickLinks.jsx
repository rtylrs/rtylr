import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faChartPie,
  faCoffee,
  faCreditCard,
  faUtensils,
} from "@fortawesome/free-solid-svg-icons"; // Example icon
import PropTypes from "prop-types";
import { Link } from "react-router-dom";

function QuickLinks() {
  return (
    <div className="quick-links">
      <QuickLink icon={faUtensils} title="Menu" />
      <QuickLink icon={faChartPie} title="Reports" />
      <QuickLink icon={faCreditCard} title="Billing" />
      <QuickLink icon={faCoffee} title="Link 4" />
      <QuickLink icon={faCoffee} title="Link 5" />
      <QuickLink icon={faCoffee} title="Link 6" />
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
