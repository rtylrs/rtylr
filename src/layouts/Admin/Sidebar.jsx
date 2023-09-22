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

const SideBar = ({ sidebarOpen, closeSidebar }) => {
  return (
    <div className={`sidebar ${sidebarOpen ? "sidebar-responsive" : ""}`}>
      <div className="sidebar__nav">
        <div className="sidebar__title__container">
          <p className="sidebar__title">Sanity Admin</p>
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
          <Link to="/admin" id="link">
            My Dashboard
          </Link>
        </div>
        <h2>Management</h2>
        <div className="sidebar__link">
          <FontAwesomeIcon icon={faUser} className="sidebar__icon" />
          <Link to="/admin/users" id="link">
            Users
          </Link>
        </div>
        <div className="sidebar__link">
          <FontAwesomeIcon icon={faBuilding} className="sidebar__icon" />
          <Link to="/admin/vendors" id="link">
            Suppliers
          </Link>
        </div>
        <div className="sidebar__link">
          <FontAwesomeIcon icon={faUserSecret} className="sidebar__icon" />
          <Link to="/admin/managment" id="link">
            Add/Remove Admins
          </Link>
        </div>

        <div className="sidebar__link">
          <FontAwesomeIcon icon={faUser} className="sidebar__icon" />
          <Link to="/admin/reviews" id="link">
            Feedback
          </Link>
        </div>
        <div className="sidebar__link">
          <FontAwesomeIcon icon={faBlog} className="sidebar__icon" />
          <Link to="/admin/reviews" id="link">
            Blog
          </Link>
        </div>
        <h2>Configuration</h2>
      </div>
    </div>
  );
};

SideBar.propTypes = {
  sidebarOpen: PropTypes.bool.isRequired,
  closeSidebar: PropTypes.func.isRequired,
};

export default SideBar;
