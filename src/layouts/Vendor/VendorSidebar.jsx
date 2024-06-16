import { useNavigate } from "react-router-dom";
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
  faUserLock,
  faUtensils,
  faMagicWandSparkles,
  faStar,
  faUserPen,
  faBullhorn,
  faCreditCard,
} from "@fortawesome/free-solid-svg-icons";
import { useSidebar } from "../../HOCs/SidebarProvider";
import { useSelector } from "react-redux";
import PackageJson from "../../../package.json";

const VendorSidebar = ({ sidebarOpen, closeSidebar }) => {
  const { managementVisible, reportsVisible, toggleSection } = useSidebar();
  const userId = useSelector((state) => state.user.userId);
  const navigate = useNavigate();

  return (
    <div
      className={`vendor-sidebar ${
        sidebarOpen ? "vendor-sidebar-responsive" : ""
      }`}
    >
      <div className="vendor-sidebar__nav">
        <div className="vendor-sidebar__title__container">
          <img
            src="/logo.png"
            alt="logo"
            className="vendor-sidebar__logo"
            onClick={() => navigate("/app/restaurant/")}
            style={{ cursor: "pointer", height: "70px", marginBottom: "10px" }}
          />
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
          onClick={() => navigate("/app/restaurant/")}
        >
          <FontAwesomeIcon icon={faHome} className="vendor-sidebar__icon" />
          <p id="link">Home</p>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/restaurant/menu" ? "vendor-active" : ""
          }`}
          onClick={() => navigate("/app/restaurant/menu")}
        >
          <FontAwesomeIcon icon={faUtensils} className="vendor-sidebar__icon" />
          <p id="link">Menu</p>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/restaurant/design"
              ? "vendor-active"
              : ""
          }`}
          onClick={() => navigate("/app/restaurant/design")}
        >
          <FontAwesomeIcon
            icon={faMagicWandSparkles}
            className="vendor-sidebar__icon"
          />
          <p id="link">Design</p>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/restaurant/qr-codes"
              ? "vendor-active"
              : ""
          }`}
          onClick={() => navigate("/app/restaurant/qr-codes")}
        >
          <FontAwesomeIcon icon={faQrcode} className="vendor-sidebar__icon" />
          <p id="link">Qr Codes</p>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/manage/design" ? "vendor-active" : ""
          }`}
          onClick={() => navigate("/manage/design")}
        >
          <FontAwesomeIcon icon={faStar} className="vendor-sidebar__icon" />
          <p id="link">Feedback</p>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/manage/design" ? "vendor-active" : ""
          }`}
          onClick={() => navigate("/manage/design")}
        >
          <FontAwesomeIcon icon={faUserPen} className="vendor-sidebar__icon" />
          <p id="link">Restaurant Information</p>
        </div>
        {/* <div
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
                onClick={() => navigate("/manage/managment")}
              >
                <FontAwesomeIcon
                  icon={faUserLock}
                  className="vendor-sidebar__icon"
                />
                <p id="link">Permissions</p>
              </div>
              <div
                className={`vendor-sidebar__link admin__link ${
                  location.pathname === "/manage/feedback"
                    ? "vendor-active"
                    : ""
                }`}
                onClick={() => navigate("/manage/feedback")}
              >
                <FontAwesomeIcon
                  icon={faRss}
                  className="vendor-sidebar__icon"
                />
                <p id="link">Feedback</p>
              </div>
            </>
          )}
        </div> */}
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
            location.pathname === "/app/restauran/billing"
              ? "vendor-active"
              : ""
          }`}
          onClick={() => navigate("/app/restaurant/billing")}
        >
          <FontAwesomeIcon
            icon={faCreditCard}
            className="vendor-sidebar__icon"
          />
          <p id="link">Billing</p>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/restauran/settings"
              ? "vendor-active"
              : ""
          }`}
          onClick={() => navigate("/app/restaurant/settings")}
        >
          <FontAwesomeIcon icon={faCog} className="vendor-sidebar__icon" />
          <p id="link">Settings</p>
        </div>
        <div
          className={`vendor-sidebar__link`}
          onClick={() => window.open(`/restaurant/${userId}`, "_blank")}
        >
          <FontAwesomeIcon
            icon={faArrowUpRightFromSquare}
            className="vendor-sidebar__icon"
          />
          <p id="link" target="_blank">
            View Menu
          </p>
        </div>
      </div>

      {/* <div
        className="vendor-sidebar__version"
        onClick={() => {
          window.open("https://voxire.com", "_blank");
        }}
      >
        <p>Version {PackageJson.version}</p>
      </div> */}
    </div>
  );
};

VendorSidebar.propTypes = {
  sidebarOpen: PropTypes.bool.isRequired,
  closeSidebar: PropTypes.func.isRequired,
};

export default VendorSidebar;
