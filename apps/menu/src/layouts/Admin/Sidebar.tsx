import { useLocation, useNavigate } from "react-router-dom";
import {
  FaTimes,
  FaHome,
  FaUser,
  FaBuilding,
  FaCube,
  FaChevronDown,
  FaChevronUp,
  FaCog,
  FaChartBar,
  FaSquare,
  FaNewspaper,
  FaRuler,
  FaCrosshairs,
  FaHandHoldingUsd,
  FaRss,
  FaPenAlt,
  FaCalendarAlt,
  FaPause,
  FaCheckSquare,
  FaChartLine,
  FaUserLock,
  FaInfoCircle,
  FaQuestionCircle,
} from "react-icons/fa";
import { useSidebar } from "../../HOCs/SidebarProvider";

const SideBar = ({ sidebarOpen, closeSidebar }) => {
  const location = useLocation();
  const navigate = useNavigate();
  const { sectionsVisible, toggleSection } = useSidebar();

  const handleNavigation = (path) => {
    navigate(path);
  };

  return (
    <div className={`sidebar ${sidebarOpen ? "sidebar-responsive" : ""}`}>
      <div className="sidebar__nav">
        <div className="sidebar__title__container">
          <p className="sidebar__title">Meow in Meows</p>
        </div>
        <FaTimes
          className="close__icon"
          onClick={closeSidebar}
          size="lg"
        />
      </div>
      <div className="sidebar__menu">
        <div
          className={`sidebar__link ${
            location.pathname === "/manage" ? "active" : ""
          }`}
          onClick={() => handleNavigation("/manage")}
        >
          <FaHome className="sidebar__icon" />
          <p id="link">Home</p>
        </div>
        <div
          className="sidebar__section"
          onClick={() => toggleSection("customers")}
        >
          <h2>
            <div>
              <FaUser className="sidebar__icon" />
              Customers
            </div>
            {sectionsVisible["customers"] ? (
              <FaChevronUp className="sidebar__icon" />
            ) : (
              <FaChevronDown className="sidebar__icon" />
            )}
          </h2>
          {sectionsVisible["customers"] && (
            <>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/users" ? "active" : ""
                }`}
                onClick={() => handleNavigation("/manage/users")}
              >
                <FaUser className="sidebar__icon" />
                <p id="link">Users</p>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/vendors" ? "active" : ""
                }`}
                onClick={() => handleNavigation("/manage/vendors")}
              >
                <FaBuilding className="sidebar__icon" />
                <p id="link">Brands</p>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/vendors" ? "active" : ""
                }`}
                onClick={() => handleNavigation("/manage/vendors")}
              >
                <FaChartLine
                  className="sidebar__icon"
                />
                <p id="link">Survey</p>
              </div>
            </>
          )}
        </div>
        <div
          className="sidebar__section"
          onClick={() => toggleSection("management")}
        >
          <h2>
            <div>
              <FaChartBar
                className="sidebar__icon"
              />
              Management
            </div>
            {sectionsVisible["management"] ? (
              <FaChevronUp className="sidebar__icon" />
            ) : (
              <FaChevronDown className="sidebar__icon" />
            )}
          </h2>
          {sectionsVisible["management"] && (
            <>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/managment" ? "active" : ""
                }`}
                onClick={() => handleNavigation("/manage/managment")}
              >
                <FaUserLock className="sidebar__icon" />
                <p id="link">Permissions</p>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/feedback" ? "active" : ""
                }`}
                onClick={() => handleNavigation("/manage/feedback")}
              >
                <FaRss className="sidebar__icon" />
                <p id="link">Feedback</p>
              </div>
            </>
          )}
        </div>
        <div
          className="sidebar__section"
          onClick={() => toggleSection("configuration")}
        >
          <h2>
            <div>
              <FaPenAlt className="sidebar__icon" />
              Sections
            </div>
            {sectionsVisible["configuration"] ? (
              <FaChevronUp className="sidebar__icon" />
            ) : (
              <FaChevronDown className="sidebar__icon" />
            )}
          </h2>
          {sectionsVisible["configuration"] && (
            <>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/sections/faqs" ? "active" : ""
                }`}
                onClick={() => handleNavigation("/manage/sections/faqs")}
              >
                <FaQuestionCircle
                  className="sidebar__icon"
                />
                <p id="link">FAQs</p>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/config/about" ? "active" : ""
                }`}
                onClick={() => handleNavigation("/manage/config/about")}
              >
                <FaInfoCircle
                  className="sidebar__icon"
                />
                <p id="link">About Us</p>
              </div>
            </>
          )}
        </div>
        <div className="sidebar__section" onClick={() => toggleSection("blog")}>
          <h2>
            <div>
              <FaCube className="sidebar__icon" />
              Blog
            </div>
            {sectionsVisible["blog"] ? (
              <FaChevronUp className="sidebar__icon" />
            ) : (
              <FaChevronDown className="sidebar__icon" />
            )}
          </h2>
          {sectionsVisible["blog"] && (
            <>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/blog" ? "active" : ""
                }`}
                onClick={() => handleNavigation("/manage/blog")}
              >
                <FaSquare className="sidebar__icon" />
                <p id="link">Blog Posts</p>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/blog" ? "active" : ""
                }`}
                onClick={() => handleNavigation("/manage/blog")}
              >
                <FaSquare className="sidebar__icon" />
                <p id="link">Submitted Posts</p>
              </div>
            </>
          )}
        </div>
        <div
          className="sidebar__section"
          onClick={() => toggleSection("reports")}
        >
          <h2>
            <div>
              <FaNewspaper className="sidebar__icon" />
              Reports
            </div>
            {sectionsVisible["reports"] ? (
              <FaChevronUp className="sidebar__icon" />
            ) : (
              <FaChevronDown className="sidebar__icon" />
            )}
          </h2>
          {sectionsVisible["reports"] && <>{/* Add report links here */}</>}
        </div>
        <div
          className="sidebar__section"
          onClick={() => toggleSection("orders")}
        >
          <h2>
            <div>
              <FaRuler className="sidebar__icon" />
              Orders
            </div>
            {sectionsVisible["orders"] ? (
              <FaChevronUp className="sidebar__icon" />
            ) : (
              <FaChevronDown className="sidebar__icon" />
            )}
          </h2>
          {sectionsVisible["orders"] && (
            <>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/orders/pending" ? "active" : ""
                }`}
                onClick={() => handleNavigation("/manage/orders/pending")}
              >
                <FaPause
                  className="sidebar__icon"
                />
                <p id="link">Pending</p>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/orders/completed"
                    ? "active"
                    : ""
                }`}
                onClick={() => handleNavigation("/manage/orders/completed")}
              >
                <FaCheckSquare
                  className="sidebar__icon"
                />
                <p id="link">Completed</p>
              </div>
            </>
          )}
        </div>
        <div
          className="sidebar__section"
          onClick={() => toggleSection("donations")}
        >
          <h2>
            <div>
              <FaHandHoldingUsd
                className="sidebar__icon"
              />
              Donations
            </div>
            {sectionsVisible["donations"] ? (
              <FaChevronUp className="sidebar__icon" />
            ) : (
              <FaChevronDown className="sidebar__icon" />
            )}
          </h2>
          {sectionsVisible["donations"] && <>{/* Add donation links here */}</>}
        </div>
        <div
          className={`sidebar__link ${
            location.pathname === "/manage/petitions" ? "active" : ""
          }`}
          onClick={() => handleNavigation("/manage/petitions")}
        >
          <FaCrosshairs className="sidebar__icon" />
          <p id="link">Petitions</p>
        </div>
        <div
          className={`sidebar__link ${
            location.pathname === "/manage/events" ? "active" : ""
          }`}
          onClick={() => handleNavigation("/manage/events")}
        >
          <FaCalendarAlt className="sidebar__icon" />
          <p id="link">Events</p>
        </div>
        <div
          className={`sidebar__link ${
            location.pathname === "/manage/settings" ? "active" : ""
          }`}
          onClick={() => handleNavigation("/manage/settings")}
        >
          <FaCog className="sidebar__icon" />
          <p id="link">Settings</p>
        </div>
      </div>
    </div>
  );
};

export default SideBar;
