import { useLocation, useNavigate } from "react-router-dom";
import PropTypes from "prop-types";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faClose,
  faHome,
  faUser,
  faBuilding,
  faCube,
  faChevronDown,
  faChevronUp,
  faShoppingBag,
  faCog,
  faBarsProgress,
  faSquare,
  faNewspaper,
  faRuler,
  faCrosshairs,
  faHandHoldingDollar,
  faRss,
  faShop,
  faPenRuler,
  faCalendarAlt,
  faCirclePause,
  faSquareCheck,
  faSquarePollVertical,
  faUserLock,
  faCircleInfo,
  faPersonCircleQuestion,
} from "@fortawesome/free-solid-svg-icons";
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
        <FontAwesomeIcon
          className="close__icon"
          onClick={closeSidebar}
          icon={faClose}
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
          <FontAwesomeIcon icon={faHome} className="sidebar__icon" />
          <p id="link">Home</p>
        </div>
        <div
          className="sidebar__section"
          onClick={() => toggleSection("customers")}
        >
          <h2>
            <div>
              <FontAwesomeIcon icon={faUser} className="sidebar__icon" />
              Customers
            </div>
            {sectionsVisible["customers"] ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
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
                <FontAwesomeIcon icon={faUser} className="sidebar__icon" />
                <p id="link">Users</p>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/vendors" ? "active" : ""
                }`}
                onClick={() => handleNavigation("/manage/vendors")}
              >
                <FontAwesomeIcon icon={faBuilding} className="sidebar__icon" />
                <p id="link">Brands</p>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/vendors" ? "active" : ""
                }`}
                onClick={() => handleNavigation("/manage/vendors")}
              >
                <FontAwesomeIcon
                  icon={faSquarePollVertical}
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
              <FontAwesomeIcon
                icon={faBarsProgress}
                className="sidebar__icon"
              />
              Management
            </div>
            {sectionsVisible["management"] ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
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
                <FontAwesomeIcon icon={faUserLock} className="sidebar__icon" />
                <p id="link">Permissions</p>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/feedback" ? "active" : ""
                }`}
                onClick={() => handleNavigation("/manage/feedback")}
              >
                <FontAwesomeIcon icon={faRss} className="sidebar__icon" />
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
              <FontAwesomeIcon icon={faPenRuler} className="sidebar__icon" />
              Sections
            </div>
            {sectionsVisible["configuration"] ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
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
                <FontAwesomeIcon
                  icon={faPersonCircleQuestion}
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
                <FontAwesomeIcon
                  icon={faCircleInfo}
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
              <FontAwesomeIcon icon={faCube} className="sidebar__icon" />
              Blog
            </div>
            {sectionsVisible["blog"] ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
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
                <FontAwesomeIcon icon={faSquare} className="sidebar__icon" />
                <p id="link">Blog Posts</p>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/blog" ? "active" : ""
                }`}
                onClick={() => handleNavigation("/manage/blog")}
              >
                <FontAwesomeIcon icon={faSquare} className="sidebar__icon" />
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
              <FontAwesomeIcon icon={faNewspaper} className="sidebar__icon" />
              Reports
            </div>
            {sectionsVisible["reports"] ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
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
              <FontAwesomeIcon icon={faRuler} className="sidebar__icon" />
              Orders
            </div>
            {sectionsVisible["orders"] ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
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
                <FontAwesomeIcon
                  icon={faCirclePause}
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
                <FontAwesomeIcon
                  icon={faSquareCheck}
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
              <FontAwesomeIcon
                icon={faHandHoldingDollar}
                className="sidebar__icon"
              />
              Donations
            </div>
            {sectionsVisible["donations"] ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
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
          <FontAwesomeIcon icon={faCrosshairs} className="sidebar__icon" />
          <p id="link">Petitions</p>
        </div>
        <div
          className={`sidebar__link ${
            location.pathname === "/manage/events" ? "active" : ""
          }`}
          onClick={() => handleNavigation("/manage/events")}
        >
          <FontAwesomeIcon icon={faCalendarAlt} className="sidebar__icon" />
          <p id="link">Events</p>
        </div>
        <div
          className={`sidebar__link ${
            location.pathname === "/manage/settings" ? "active" : ""
          }`}
          onClick={() => handleNavigation("/manage/settings")}
        >
          <FontAwesomeIcon icon={faCog} className="sidebar__icon" />
          <p id="link">Settings</p>
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
