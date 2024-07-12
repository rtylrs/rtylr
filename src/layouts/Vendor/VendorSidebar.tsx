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
  faQrcode,
  faUtensils,
  faMagicWandSparkles,
  faStar,
  faUserPen,
  faCreditCard,
  faPuzzlePiece,
  faFolder,
  faUtensilSpoon,
  faEye,
  faChartLine,
} from "@fortawesome/free-solid-svg-icons";
import { useSidebar } from "../../HOCs/SidebarProvider";
import { useSelector } from "react-redux";
import { RootState } from "../../redux/store";

const VendorSidebar = ({ sidebarOpen, closeSidebar }) => {
  const { menuVisible, toggleSection } = useSidebar();
  const userId = useSelector((state: RootState) => state.user.userId);
  const userAddons = useSelector((state: RootState) => state.user.addons);
  const navigate = useNavigate();
  // const user = useSelector((state: RootState) => state.user.user);

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
            onClick={() => navigate("/app/business/")}
            style={{ cursor: "pointer", height: "70px", marginBottom: "10px" }}
          />
        </div>
        <FontAwesomeIcon
          className="vendor-close__icon"
          onClick={() => closeSidebar()}
          icon={faClose}
          size="lg"
        />
      </div>
      <div className="vendor-sidebar__menu">
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/business/" ? "vendor-active" : ""
          }`}
          onClick={() => navigate("/app/business/")}
        >
          <FontAwesomeIcon icon={faHome} className="vendor-sidebar__icon" />
          <p id="link">Home</p>
        </div>
        <div
          className="vendor-sidebar__section"
          onClick={() => toggleSection("menu")}
        >
          <h2>
            <div>
              <FontAwesomeIcon
                icon={faUtensils}
                className="vendor-sidebar__icon"
              />
              Menu
            </div>
            {menuVisible ? (
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
          {menuVisible && (
            <>
              <div
                className={`vendor-sidebar__link vendor-admin__link  ${
                  location.pathname === "/app/business/menu"
                    ? "vendor-active"
                    : ""
                }`}
                onClick={() => navigate("/app/business/menu")}
              >
                <FontAwesomeIcon
                  icon={faUtensilSpoon}
                  className="vendor-sidebar__icon"
                />
                <p id="link">Catalog</p>
              </div>
              <div
                className={`vendor-sidebar__link vendor-admin__link  ${
                  location.pathname === "/app/business/design"
                    ? "vendor-active"
                    : ""
                }`}
                onClick={() => navigate("/app/business/design")}
              >
                <FontAwesomeIcon
                  icon={faMagicWandSparkles}
                  className="vendor-sidebar__icon"
                />
                <p id="link">Customize</p>
              </div>
              <div
                className={`vendor-sidebar__link vendor-admin__link  ${
                  location.pathname === "/app/business/preview"
                    ? "vendor-active"
                    : ""
                }`}
                onClick={() => navigate("/app/business/preview")}
              >
                <FontAwesomeIcon
                  icon={faEye}
                  className="vendor-sidebar__icon"
                />
                <p id="link">Preview</p>
              </div>
            </>
          )}
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/business/qr-codes"
              ? "vendor-active"
              : ""
          }`}
          onClick={() => navigate("/app/business/qr-codes")}
        >
          <FontAwesomeIcon icon={faQrcode} className="vendor-sidebar__icon" />
          <p id="link">Qr Codes</p>
        </div>
        {userAddons && userAddons?.feedback && (
          <div
            className={`vendor-sidebar__link ${
              location.pathname === "/manage/design" ? "vendor-active" : ""
            }`}
            onClick={() => navigate("/manage/design")}
          >
            <FontAwesomeIcon icon={faStar} className="vendor-sidebar__icon" />
            <p id="link">Feedback</p>
          </div>
        )}
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/business/menus" ? "vendor-active" : ""
          }`}
          onClick={() => navigate("/app/business/menus")}
        >
          <FontAwesomeIcon icon={faFolder} className="vendor-sidebar__icon" />
          <p id="link">My Menus</p>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/manage/design" ? "vendor-active" : ""
          }`}
          onClick={() => navigate("/manage/design")}
        >
          <FontAwesomeIcon icon={faUserPen} className="vendor-sidebar__icon" />
          <p id="link">Information</p>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/manage/design" ? "vendor-active" : ""
          }`}
          onClick={() => navigate("/manage/design")}
        >
          <FontAwesomeIcon
            icon={faChartLine}
            className="vendor-sidebar__icon"
          />
          <p id="link">Analytics</p>
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
          className={`vendor-sidebar__link ${
            location.pathname === "/app/business/addons" ? "vendor-active" : ""
          }`}
          onClick={() => navigate("/app/business/addons")}
        >
          <FontAwesomeIcon
            icon={faPuzzlePiece}
            className="vendor-sidebar__icon"
          />
          <p id="link"> Add-Ons</p>
        </div>
        {/* <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/business/addons"
              ? "vendor-active"
              : ""
          }`}
        >
          <FontAwesomeIcon icon={faWallet} className="vendor-sidebar__icon" />
          <Link to="/app/business/addons" id="link">
            Commissions
          </Link>
        </div> */}
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/business/billing" ? "vendor-active" : ""
          }`}
          onClick={() => navigate("/app/business/billing")}
        >
          <FontAwesomeIcon
            icon={faCreditCard}
            className="vendor-sidebar__icon"
          />
          <p id="link">Subscription</p>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/business/settings"
              ? "vendor-active"
              : ""
          }`}
          onClick={() => navigate("/app/business/settings")}
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
          <p id="link">View Menu</p>
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
