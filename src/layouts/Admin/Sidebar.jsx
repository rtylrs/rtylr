import { Link } from "react-router-dom";
import PropTypes from "prop-types";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faClose,
  faHome,
  faUser,
  faBuilding,
  faUserSecret,
  faCube,
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
          <FontAwesomeIcon icon={faUserSecret} className="sidebar__icon" />
          <Link to="/manage/managment" id="link">
            Add/Remove Admins
          </Link>
        </div>
        <div className="sidebar__link">
          <FontAwesomeIcon icon={faCube} className="sidebar__icon" />
          <Link to="/manage/blog" id="link">
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
