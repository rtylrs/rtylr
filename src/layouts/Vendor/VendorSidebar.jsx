import { Link } from "react-router-dom";
import PropTypes from "prop-types";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faClose,
  faHome,
  faChevronDown,
  faChevronUp,
  faCog,
  faArrowUpRightFromSquare,
  faNewspaper,
  faRss,
  faQrcode,
  faPuzzlePiece,
  faUserLock,
  faUtensils,
  faMagicWandSparkles,
  faStar,
  faUserPen,
  faBullhorn,
  faCreditCard,
  faWallet,
} from "@fortawesome/free-solid-svg-icons";
import { useSidebar } from "../../HOCs/SidebarProvider";
import { useSelector } from "react-redux";

const VendorSidebar = ({ sidebarOpen, closeSidebar }) => {
  const { managementVisible, reportsVisible, toggleSection } = useSidebar();
  const userId = useSelector((state) => state.user.userId);

  return (
    <div
      className={`vendor-sidebar ${sidebarOpen ? "sidebar-responsive" : ""}`}
    >
      <div className="vendor-sidebar__nav">
        <div className="vendor-sidebar__title__container">
          <p className="vendor-sidebar__title">Vbitez</p>
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
            location.pathname === "/app/restaurant/" ? "vendor-active" : ""
          }`}
        >
          <FontAwesomeIcon icon={faHome} className="vendor-sidebar__icon" />
          <Link to="/app/restaurant/" id="link">
            Home
          </Link>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/restaurant/menu" ? "vendor-active" : ""
          }`}
        >
          <FontAwesomeIcon icon={faUtensils} className="vendor-sidebar__icon" />
          <Link to="/app/restaurant/menu" id="link">
            Menu
          </Link>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/manage/design" ? "vendor-active" : ""
          }`}
        >
          <FontAwesomeIcon
            icon={faMagicWandSparkles}
            className="vendor-sidebar__icon"
          />
          <Link to="/manage/design" id="link">
            Design
          </Link>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/restaurant/qr-codes"
              ? "vendor-active"
              : ""
          }`}
        >
          <FontAwesomeIcon icon={faQrcode} className="vendor-sidebar__icon" />
          <Link to="/app/restaurant/qr-codes" id="link">
            Qr Codes
          </Link>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/manage/design" ? "vendor-active" : ""
          }`}
        >
          <FontAwesomeIcon icon={faStar} className="vendor-sidebar__icon" />
          <Link to="/manage/design" id="link">
            Feedback
          </Link>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/manage/design" ? "vendor-active" : ""
          }`}
        >
          <FontAwesomeIcon icon={faUserPen} className="vendor-sidebar__icon" />
          <Link to="/manage/design" id="link">
            Restaurant Information
          </Link>
        </div>
        <div
          className="vendor-sidebar__section"
          onClick={() => toggleSection("management")}
        >
          <h2>
            <div>
              <FontAwesomeIcon
                icon={faBullhorn}
                className="vendor-sidebar__icon"
              />
              Marketing
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
        {/* <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/restaurant/addons"
              ? "vendor-active"
              : ""
          }`}
        >
          <FontAwesomeIcon
            icon={faPuzzlePiece}
            className="vendor-sidebar__icon"
          />
          <Link to="/app/restaurant/addons" id="link">
            Add-Ons
          </Link>
        </div> */}
        {/* <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/restaurant/addons"
              ? "vendor-active"
              : ""
          }`}
        >
          <FontAwesomeIcon icon={faWallet} className="vendor-sidebar__icon" />
          <Link to="/app/restaurant/addons" id="link">
            Commissions
          </Link>
        </div> */}
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/restauran/settings"
              ? "vendor-active"
              : ""
          }`}
        >
          <FontAwesomeIcon
            icon={faCreditCard}
            className="vendor-sidebar__icon"
          />
          <Link to="/app/restaurant/settings" id="link">
            Billing
          </Link>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/restauran/settings"
              ? "vendor-active"
              : ""
          }`}
        >
          <FontAwesomeIcon icon={faCog} className="vendor-sidebar__icon" />
          <Link to="/app/restaurant/settings" id="link">
            Settings
          </Link>
        </div>
        <div className={`vendor-sidebar__link`}>
          <FontAwesomeIcon
            icon={faArrowUpRightFromSquare}
            className="vendor-sidebar__icon"
          />
          {/* TODO: Make it take them to their Menu */}
          <Link to={`/restaurant/${userId}`} id="link" target="_blank">
            View Menu
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
