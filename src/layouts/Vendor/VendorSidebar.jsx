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
  faArrowUpRightFromSquare,
  faBarsProgress,
  faSquare,
  faNewspaper,
  faRuler,
  faCrosshairs,
  faHandHoldingDollar,
  faRss,
  faWarehouse,
  faShop,
  faQrcode,
  faHeadset,
  faPuzzlePiece,
  faPenRuler,
  faCalendarAlt,
  faCirclePause,
  faSquareCheck,
  faSquarePollVertical,
  faUserLock,
  faCircleInfo,
} from "@fortawesome/free-solid-svg-icons";
import { useSidebar } from "../../HOCs/SidebarProvider";

import QRCode from "qrcode";
import { faFirstdraft } from "@fortawesome/free-brands-svg-icons";

//TODO: FIX THIS TO BE TAILORED TO VENDOR
const VendorSidebar = ({ sidebarOpen, closeSidebar }) => {
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
    <div
      className={`vendor-sidebar ${sidebarOpen ? "sidebar-responsive" : ""}`}
    >
      <div className="vendor-sidebar__nav">
        <div className="vendor-sidebar__title__container">
          <p className="vendor-sidebar__title">Vbitez</p>
          <div>
            <FontAwesomeIcon
              icon={faHeadset}
              title="Contact Support"
              onClick={() => {
                window.open("https://wa.me/message/D5ZSC3DH3NI6P1", "_blank");
              }}
            />
          </div>
        </div>
        <FontAwesomeIcon
          className="vendor-close__icon"
          onClick={() => closeSidebar()}
          icon={faClose}
          r
          size="lg"
        />
      </div>
      <div className="vendor-sidebar__menu">
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/manager/restaurant/" ? "vendor-active" : ""
          }`}
        >
          <FontAwesomeIcon icon={faHome} className="vendor-sidebar__icon" />
          <Link to="/manager/restaurant/" id="link">
            Home
          </Link>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/manage/pos" ? "vendor-active" : ""
          }`}
        >
          <FontAwesomeIcon
            icon={faWarehouse}
            className="vendor-sidebar__icon"
            style={{
              fontSize: "14px",
            }}
          />
          <Link to="/manage/pos" id="link">
            POS
          </Link>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/manage/products" ? "vendor-active" : ""
          }`}
        >
          <FontAwesomeIcon
            icon={faShoppingBag}
            className="vendor-sidebar__icon"
          />
          <Link to="/manage/products" id="link">
            Products
          </Link>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/manage/branches" ? "vendor-active" : ""
          }`}
        >
          <FontAwesomeIcon icon={faShop} className="vendor-sidebar__icon" />
          <Link to="/manage/branches" id="link">
            Branches
          </Link>
        </div>
        <div
          className="vendor-sidebar__section"
          onClick={() => toggleSection("customers")}
        >
          <h2>
            <div>
              <FontAwesomeIcon icon={faUser} className="vendor-sidebar__icon" />
              Customers
            </div>
            {customersVisible ? (
              <FontAwesomeIcon
                icon={faChevronUp}
                className="vendor-sidebar__icon"
              />
            ) : (
              <FontAwesomeIcon
                icon={faChevronDown}
                className="vendor-sidebar__icon"
              />
            )}
          </h2>
          {customersVisible && (
            <>
              <div
                className={`vendor-sidebar__link admin__link ${
                  location.pathname === "/manage/users" ? "vendor-active" : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faUser}
                  className="vendor-sidebar__icon"
                />
                <Link to="/manage/users" id="link">
                  Users
                </Link>
              </div>
              <div
                className={`vendor-sidebar__link admin__link ${
                  location.pathname === "/manage/vendors" ? "vendor-active" : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faBuilding}
                  className="vendor-sidebar__icon"
                />
                <Link to="/manage/vendors" id="link">
                  Brands
                </Link>
              </div>
              <div
                className={`vendor-sidebar__link admin__link ${
                  location.pathname === "/manage/vendors" ? "vendor-active" : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faSquarePollVertical}
                  className="vendor-sidebar__icon"
                />
                <Link to="/manage/vendors" id="link">
                  Survey
                </Link>
              </div>
            </>
          )}
        </div>
        <div
          className="vendor-sidebar__section"
          onClick={() => toggleSection("management")}
        >
          <h2>
            <div>
              <FontAwesomeIcon
                icon={faBarsProgress}
                className="vendor-sidebar__icon"
              />
              Management
            </div>
            {managementVisible ? (
              <FontAwesomeIcon
                icon={faChevronUp}
                className="vendor-sidebar__icon"
              />
            ) : (
              <FontAwesomeIcon
                icon={faChevronDown}
                className="vendor-sidebar__icon"
              />
            )}
          </h2>
          {managementVisible && (
            <>
              <div
                className={`vendor-sidebar__link admin__link ${
                  location.pathname === "/manage/managment"
                    ? "vendor-active"
                    : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faUserLock}
                  className="vendor-sidebar__icon"
                />
                <Link to="/manage/managment" id="link">
                  Permissions
                </Link>
              </div>
              <div
                className={`vendor-sidebar__link admin__link ${
                  location.pathname === "/manage/feedback"
                    ? "vendor-active"
                    : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faRss}
                  className="vendor-sidebar__icon"
                />
                <Link to="/manage/feedback" id="link">
                  Feedback
                </Link>
              </div>
            </>
          )}
        </div>
        <div
          className="vendor-sidebar__section"
          onClick={() => toggleSection("configuration")}
        >
          <h2>
            <div>
              <FontAwesomeIcon
                icon={faPenRuler}
                className="vendor-sidebar__icon"
              />
              Sections
            </div>
            {configurationVisible ? (
              <FontAwesomeIcon
                icon={faChevronUp}
                className="vendor-sidebar__icon"
              />
            ) : (
              <FontAwesomeIcon
                icon={faChevronDown}
                className="vendor-sidebar__icon"
              />
            )}
          </h2>
          {configurationVisible && (
            <>
              <div
                className={`vendor-sidebar__link admin__link ${
                  location.pathname === "/manage/config/home"
                    ? "vendor-active"
                    : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faFirstdraft}
                  className="vendor-sidebar__icon"
                />
                <Link to="/manage/config/home" id="link">
                  Home
                </Link>
              </div>
              <div
                className={`vendor-sidebar__link admin__link ${
                  location.pathname === "/manage/config/about"
                    ? "vendor-active"
                    : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faCircleInfo}
                  className="vendor-sidebar__icon"
                />
                <Link to="/manage/config/about" id="link">
                  About Us
                </Link>
              </div>
            </>
          )}
        </div>
        <div
          className="vendor-sidebar__section"
          onClick={() => toggleSection("blog")}
        >
          <h2>
            <div>
              <FontAwesomeIcon icon={faCube} className="vendor-sidebar__icon" />
              Blog
            </div>
            {blogVisible ? (
              <FontAwesomeIcon
                icon={faChevronUp}
                className="vendor-sidebar__icon"
              />
            ) : (
              <FontAwesomeIcon
                icon={faChevronDown}
                className="vendor-sidebar__icon"
              />
            )}
          </h2>
          {blogVisible && (
            <>
              <div
                className={`vendor-sidebar__link admin__link ${
                  location.pathname === "/manage/blog" ? "vendor-active" : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faSquare}
                  className="vendor-sidebar__icon"
                />
                <Link to="/manage/blog" id="link">
                  Blog Posts
                </Link>
              </div>
              <div
                className={`vendor-sidebar__link admin__link ${
                  location.pathname === "/manage/blog" ? "vendor-active" : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faSquare}
                  className="vendor-sidebar__icon"
                />
                <Link to="/manage/blog" id="link">
                  Submited Posts
                </Link>
              </div>
            </>
          )}
        </div>
        <div
          className="vendor-sidebar__section"
          onClick={() => toggleSection("reports")}
        >
          <h2>
            <div>
              <FontAwesomeIcon
                icon={faNewspaper}
                className="vendor-sidebar__icon"
              />
              Reports
            </div>
            {reportsVisible ? (
              <FontAwesomeIcon
                icon={faChevronUp}
                className="vendor-sidebar__icon"
              />
            ) : (
              <FontAwesomeIcon
                icon={faChevronDown}
                className="vendor-sidebar__icon"
              />
            )}
          </h2>
          {reportsVisible && <>{/* Add report links here */}</>}
        </div>
        <div
          className="vendor-sidebar__section"
          onClick={() => toggleSection("orders")}
        >
          <h2>
            <div>
              <FontAwesomeIcon
                icon={faRuler}
                className="vendor-sidebar__icon"
              />
              Orders
            </div>
            {ordersVisible ? (
              <FontAwesomeIcon
                icon={faChevronUp}
                className="vendor-sidebar__icon"
              />
            ) : (
              <FontAwesomeIcon
                icon={faChevronDown}
                className="vendor-sidebar__icon"
              />
            )}
          </h2>
          {ordersVisible && (
            <>
              <div
                className={`vendor-sidebar__link admin__link ${
                  location.pathname === "/manage/orders/pending"
                    ? "vendor-active"
                    : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faCirclePause}
                  className="vendor-sidebar__icon"
                />
                <Link to="/manage/orders/pending" id="link">
                  Pending
                </Link>
              </div>
              <div
                className={`vendor-sidebar__link admin__link ${
                  location.pathname === "/manage/orders/completed"
                    ? "vendor-active"
                    : ""
                }`}
              >
                <FontAwesomeIcon
                  icon={faSquareCheck}
                  className="vendor-sidebar__icon"
                />
                <Link to="/manage/orders/completed" id="link">
                  Completed
                </Link>
              </div>
            </>
          )}
        </div>
        <div
          className="vendor-sidebar__section"
          onClick={() => toggleSection("donations")}
        >
          <h2>
            <div>
              <FontAwesomeIcon
                icon={faHandHoldingDollar}
                className="vendor-sidebar__icon"
              />
              Donations
            </div>
            {donationsVisible ? (
              <FontAwesomeIcon
                icon={faChevronUp}
                className="vendor-sidebar__icon"
              />
            ) : (
              <FontAwesomeIcon
                icon={faChevronDown}
                className="vendor-sidebar__icon"
              />
            )}
          </h2>
          {donationsVisible && <>{/* Add donation links here */}</>}
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/manage/events" ? "vendor-active" : ""
          }`}
        >
          <FontAwesomeIcon
            icon={faCalendarAlt}
            className="vendor-sidebar__icon"
          />
          <Link to="/manage/events" id="link">
            Events
          </Link>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/manager/restaurant/manage/addons"
              ? "vendor-active"
              : ""
          }`}
        >
          <FontAwesomeIcon
            icon={faPuzzlePiece}
            className="vendor-sidebar__icon"
          />
          <Link to="/owner/restaurant/manage/addons" id="link">
            Add-Ons
          </Link>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/manager/restauran/settings"
              ? "vendor-active"
              : ""
          }`}
        >
          <FontAwesomeIcon icon={faCog} className="vendor-sidebar__icon" />
          <Link to="/manager/restaurant/settings" id="link">
            Settings
          </Link>
        </div>
        <div className={`vendor-sidebar__link`}>
          <FontAwesomeIcon
            icon={faArrowUpRightFromSquare}
            className="vendor-sidebar__icon"
          />
          {/* TODO: Make it take them to their Menu */}
          <Link to="/restaurant/69" id="link" target="_blank">
            Open Menu
          </Link>
        </div>
      </div>
    </div>
  );
};

VendorSidebar.propTypes = {
  sidebarOpen: PropTypes.bool.isRequired,
  closeSidebar: PropTypes.func.isRequired,
};

export default VendorSidebar;
