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
  faChevronDown,
  faChevronUp,
  faShoppingBag,
  faCog,
  faArrowUpRightFromSquare,
  faBarsProgress,
  faCheckDouble,
  faSquare,
  faNewspaper,
  faRuler,
  faCrosshairs,
  faHandHoldingDollar,
  faRss,
  faWarehouse,
  faShop,
} from "@fortawesome/free-solid-svg-icons";
import { useSidebar } from "../../HOCs/SidebarProvider";

const SideBar = ({ sidebarOpen, closeSidebar }) => {
  // Define state to track the visibility of each section
  const {
    customersVisible,
    managementVisible,
    configurationVisible,
    blogVisible,
    reportsVisible,
    ordersVisible,
    donationsVisible,
    toggleSection,
  } = useSidebar(); // Access the context using the hook

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
        <div className="sidebar__link">
          <FontAwesomeIcon icon={faHome} className="sidebar__icon" />
          <Link to="/manage" id="link">
            My Dashboard
          </Link>
        </div>
        <div className="sidebar__link">
          <FontAwesomeIcon icon={faWarehouse} className="sidebar__icon" />
          <Link to="/manage/products" id="link">
            POS
          </Link>
        </div>
        <div className="sidebar__link">
          <FontAwesomeIcon icon={faShoppingBag} className="sidebar__icon" />
          <Link to="/manage/products" id="link">
            Sanity Products
          </Link>
        </div>
        <div className="sidebar__link">
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
              <div className="sidebar__link  admin__link">
                <FontAwesomeIcon icon={faUser} className="sidebar__icon" />
                <Link to="/manage/users" id="link">
                  Users
                </Link>
              </div>
              <div className="sidebar__link  admin__link">
                <FontAwesomeIcon icon={faBuilding} className="sidebar__icon" />
                <Link to="/manage/vendors" id="link">
                  Brands
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
              <div className="sidebar__link  admin__link">
                <FontAwesomeIcon
                  icon={faUserSecret}
                  className="sidebar__icon"
                />
                <Link to="/manage/managment" id="link">
                  Permissions
                </Link>
              </div>
              <div className="sidebar__link  admin__link">
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
              <FontAwesomeIcon icon={faCheckDouble} className="sidebar__icon" />
              Configuration
            </div>
            {configurationVisible ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
            )}
          </h2>
          {configurationVisible && <>{/* Add configuration links here */}</>}
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
              <div className="sidebar__link  admin__link">
                <FontAwesomeIcon icon={faSquare} className="sidebar__icon" />
                <Link to="/manage/blog" id="link">
                  Blog Posts
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
          {ordersVisible && <>{/* Add order links here */}</>}
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
        <div className="sidebar__link ">
          <FontAwesomeIcon icon={faCrosshairs} className="sidebar__icon" />
          <Link to="/manage/petitions" id="link">
            Petitions
          </Link>
        </div>
        <div className="sidebar__link">
          <FontAwesomeIcon icon={faCog} className="sidebar__icon" />
          <Link to="/manage/settings" id="link">
            Settings
          </Link>
        </div>
        <div className="sidebar__link">
          <FontAwesomeIcon
            icon={faArrowUpRightFromSquare}
            className="sidebar__icon"
          />
          <Link to="/" id="link" target="_blank">
            Open Website
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
