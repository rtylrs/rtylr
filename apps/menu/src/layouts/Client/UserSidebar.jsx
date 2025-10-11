import { Link } from "react-router-dom";
import {
  FaTimes,
  FaHome,
  FaUser,
  FaHeart,
  FaCog,
  FaHistory,
  FaCube,
} from "react-icons/fa";

const SideBar = ({ sidebarOpen, closeSidebar }) => {
  return (
    <div className={`sidebar ${sidebarOpen ? "sidebar-responsive" : ""}`}>
      <div className="sidebar__nav">
        <div className="logo">
          {/* <Link to="/">
            <img src={Logo} alt="logo" className="sidebar__logo" />
          </Link> */}
        </div>
        <FaTimes
          className="close__icon"
          onClick={() => closeSidebar()}
          size="lg"
        />
      </div>
      <div className="sidebar__menu">
        <div className="sidebar__link ">
          <FaHome className="sidebar__icon" />
          <Link to="/account" id="link">
            Dashboard
          </Link>
        </div>
        <h2>Management</h2>
        <div className="sidebar__link">
          <FaUser className="sidebar__icon" />
          <Link to="/profile" id="link">
            Profile
          </Link>
        </div>
        <div className="sidebar__link">
          <FaHeart className="sidebar__icon" />
          <Link to="/favorites" id="link">
            Favorites
          </Link>
        </div>
        <div className="sidebar__link">
          <FaHistory className="sidebar__icon" />
          <Link to="/my-history" id="link">
            Previous Orders
          </Link>
        </div>
        <div className="sidebar__link">
          <FaCube className="sidebar__icon" />
          <Link to="/blog/public/add" id="link">
            Submit a Blog
          </Link>
        </div>
        <div className="sidebar__link">
          <FaCog className="sidebar__icon" />
          <Link to="/settings" id="link">
            Settings
          </Link>
        </div>
      </div>
    </div>
  );
};

export default SideBar;
