import { Link } from "react-router-dom";
import PropTypes from "prop-types";
import { faClose } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

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
          <i className="fa fa-home"></i>
          <Link to="/admin" id="link">
            My Dashboard
          </Link>
        </div>
        <h2>Management</h2>
        <div className="sidebar__link">
          <i className="fa fa-user"></i>
          <Link to="/admin/users" id="link">
            Users
          </Link>
        </div>
        <div className="sidebar__link">
          <i className="fa-solid fa-building"></i>
          <Link to="/admin/vendors" id="link">
            Suppliers
          </Link>
        </div>
        <div className="sidebar__link">
          <i className="fa fa-user-secret"></i>
          <Link to="/admin/managment" id="link">
            Add/Remove Admins
          </Link>
        </div>

        <div className="sidebar__link">
          <i className="fa-solid fa-heart"></i>
          <Link to="/admin/reviews" id="link">
            Feedback
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
