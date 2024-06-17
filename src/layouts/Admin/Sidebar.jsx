import { Link } from "react-router-dom";
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
  const {
    customersVisible,
    managementVisible,
    configurationVisible,
    blogVisible,
    reportsVisible,
    ordersVisible,
    donationsVisible,
    toggleSection,
  } = useSidebar();

  return (
    <div className={`sidebar ${sidebarOpen ? "sidebar-responsive" : ""}`}>
      <div className="sidebar__nav">
        <div className="sidebar__title__container">
          <p className="sidebar__title">Vbitez Admin</p>
        </div>
        <FontAwesomeIcon
          className="close__icon"
          onClick={() => closeSidebar()}
          icon={faClose}
          r
          size="lg"
        />
      </div>
      <div className="sidebar__menu">
        <div
          className={`sidebar__link ${
            location.pathname === "/manage" ? "active" : ""
          }`}
        >
          <FontAwesomeIcon icon={faHome} className="sidebar__icon" />
          <Link to="/manage" id="link">
            Home
          </Link>
        </div>
        <div
          className={`sidebar__link ${
            location.pathname === "/manage/products" ? "active" : ""
          }`}
        >
          <FontAwesomeIcon icon={faShoppingBag} className="sidebar__icon" />
          <Link to="/manage/products" id="link">
            Products
          </Link>
        </div>
        <div
          className={`sidebar__link ${
            location.pathname === "/manage/branches" ? "active" : ""
          }`}
        >
          <FontAwesomeIcon icon={faShop} className="sidebar__icon" />
          <Link to="/manage/branches" id="link">
            Branches
          </Link>
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
            {customersVisible ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
            )}
          </h2>
          {customersVisible && (
            <>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/users" ? "active" : ""
                }`}
              >
                <FontAwesomeIcon icon={faUser} className="sidebar__icon" />
                <Link to="/manage/users" id="link">
                  Users
                </Link>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/vendors" ? "active" : ""
                }`}
              >
                <FontAwesomeIcon icon={faBuilding} className="sidebar__icon" />
                <Link to="/manage/vendors" id="link">
                  Brands
                </Link>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/vendors" ? "active" : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faSquarePollVertical}
                  className="sidebar__icon"
                />
                <Link to="/manage/vendors" id="link">
                  Survey
                </Link>
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
            {managementVisible ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
            )}
          </h2>
          {managementVisible && (
            <>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/managment" ? "active" : ""
                }`}
              >
                <FontAwesomeIcon icon={faUserLock} className="sidebar__icon" />
                <Link to="/manage/managment" id="link">
                  Permissions
                </Link>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/feedback" ? "active" : ""
                }`}
              >
                <FontAwesomeIcon icon={faRss} className="sidebar__icon" />
                <Link to="/manage/feedback" id="link">
                  Feedback
                </Link>
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
            {configurationVisible ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
            )}
          </h2>
          {configurationVisible && (
            <>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/sections/faqs" ? "active" : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faPersonCircleQuestion}
                  className="sidebar__icon"
                />
                <Link to="/manage/sections/faqs" id="link">
                  FAQs
                </Link>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/config/about" ? "active" : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faCircleInfo}
                  className="sidebar__icon"
                />
                <Link to="/manage/config/about" id="link">
                  About Us
                </Link>
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
            {blogVisible ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
            )}
          </h2>
          {blogVisible && (
            <>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/blog" ? "active" : ""
                }`}
              >
                <FontAwesomeIcon icon={faSquare} className="sidebar__icon" />
                <Link to="/manage/blog" id="link">
                  Blog Posts
                </Link>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/blog" ? "active" : ""
                }`}
              >
                <FontAwesomeIcon icon={faSquare} className="sidebar__icon" />
                <Link to="/manage/blog" id="link">
                  Submited Posts
                </Link>
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
            {reportsVisible ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
            )}
          </h2>
          {reportsVisible && <>{/* Add report links here */}</>}
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
            {ordersVisible ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
            )}
          </h2>
          {ordersVisible && (
            <>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/orders/pending" ? "active" : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faCirclePause}
                  className="sidebar__icon"
                />
                <Link to="/manage/orders/pending" id="link">
                  Pending
                </Link>
              </div>
              <div
                className={`sidebar__link admin__link ${
                  location.pathname === "/manage/orders/completed"
                    ? "active"
                    : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faSquareCheck}
                  className="sidebar__icon"
                />
                <Link to="/manage/orders/completed" id="link">
                  Completed
                </Link>
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
            {donationsVisible ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
            )}
          </h2>
          {donationsVisible && <>{/* Add donation links here */}</>}
        </div>
        <div
          className={`sidebar__link ${
            location.pathname === "/manage/petitions" ? "active" : ""
          }`}
        >
          <FontAwesomeIcon icon={faCrosshairs} className="sidebar__icon" />
          <Link to="/manage/petitions" id="link">
            Petitions
          </Link>
        </div>
        <div
          className={`sidebar__link ${
            location.pathname === "/manage/events" ? "active" : ""
          }`}
        >
          <FontAwesomeIcon icon={faCalendarAlt} className="sidebar__icon" />
          <Link to="/manage/events" id="link">
            Events
          </Link>
        </div>
        <div
          className={`sidebar__link ${
            location.pathname === "/manage/settings" ? "active" : ""
          }`}
        >
          <FontAwesomeIcon icon={faCog} className="sidebar__icon" />
          <Link to="/manage/settings" id="link">
            Settings
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
