import { Link } from "react-router-dom";
import PropTypes from "prop-types";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faClose,
  faHome,
  faUser,
  faBuilding,
  faUserSecret,
  faBlog,
} from "@fortawesome/free-solid-svg-icons";
import Logo from "../../assets/logo.png";

const SideBar = ({ sidebarOpen, closeSidebar }) => {
  return (
    <div className={`sidebar ${sidebarOpen ? "sidebar-responsive" : ""}`}>
      <div className="sidebar__nav">
        <div className="logo">
          <Link to="/">
            <img src={Logo} alt="logo" className="sidebar__logo" />
          </Link>
        </div>
        <FontAwesomeIcon
          className="close__icon"
          onClick={() => closeSidebar()}
          icon={faClose}
          size="lg"
        />
      </div>
      <div className="sidebar__menu">
        <div className="sidebar__link ">
          <FontAwesomeIcon icon={faHome} className="sidebar__icon" />
          <Link to="/account" id="link">
            My Dashboard
          </Link>
        </div>
        <h2>Management</h2>
        <div className="sidebar__link">
          <FontAwesomeIcon icon={faUser} className="sidebar__icon" />
          <Link to="/profile" id="link">
            Profile
          </Link>
        </div>
        <div className="sidebar__link">
          <FontAwesomeIcon icon={faBuilding} className="sidebar__icon" />
          <Link to="/favorites" id="link">
            Favorites
          </Link>
        </div>
        <div className="sidebar__link">
          <FontAwesomeIcon icon={faUser} className="sidebar__icon" />
          <Link to="/my-history" id="link">
            Previous Orders
          </Link>
        </div>
        <div className="sidebar__link">
          <FontAwesomeIcon icon={faUserSecret} className="sidebar__icon" />
          <Link to="/settings" id="link">
            Settings
          </Link>
        </div>

        <div className="sidebar__link">
          <FontAwesomeIcon icon={faBlog} className="sidebar__icon" />
          <Link to="/join" id="link">
            Become a business
          </Link>
        </div>
      </div>
    </div>
  );
};

SideBar.propTypes = {
  sidebarOpen: PropTypes.bool.isRequired,
  closeSidebar: PropTypes.func.isRequired,
};

export default SideBar;
