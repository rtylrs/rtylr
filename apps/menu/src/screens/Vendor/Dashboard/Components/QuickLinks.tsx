import {
  FaBookOpen,
  FaCog,
  FaCreditCard,
  FaFolder,
  FaMagic,
  FaPhone,
  FaPlus,
  FaQrcode,
} from "react-icons/fa";
import { Link } from "react-router-dom";

function QuickLinks() {
  return (
    <div className="quick-links">
      <QuickLink icon={FaBookOpen} title="Menu" to="/app/business/menu" />
      <QuickLink icon={FaFolder} title="My Menus" to="/app/business/menus" />
      {/* <QuickLink icon={faChartPie} title="Reports" to="/app/business/reports" /> */}
      <QuickLink
        icon={FaQrcode}
        title="Qr Codes"
        to="/app/business/qr-codes/"
      />
      <QuickLink
        icon={FaPlus}
        title="Create Qr"
        to="/app/business/qr-codes/create"
      />
      <QuickLink
        icon={FaMagic}
        title="Appearance"
        to="/app/business/design"
      />
      <QuickLink
        icon={FaCreditCard}
        title="Billing"
        to="/app/business/billing"
      />
      <QuickLink icon={FaCog} title="Settings" to="/app/business/settings" />
      <QuickLink icon={FaPhone} title="Support" to="/contact" />
    </div>
  );
}

const QuickLink = ({ icon, title, to }) => (
  <Link className="quick-link" to={to}>
    <icon />
    <span>{title}</span>
  </Link>
);

export default QuickLinks;
