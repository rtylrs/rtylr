import React from "react";
import { useNavigate } from "react-router-dom";
import PropTypes from "prop-types";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faClose,
  faHome,
  // faChevronDown,
  // faChevronUp,
  faCog,
  faArrowUpRightFromSquare,
  faQrcode,
  faMagicWandSparkles,
  faStar,
  faUserPen,
  faCreditCard,
  faPuzzlePiece,
  faFolder,
  faChartLine,
  faGlobe,
  // faUtensils,
  faBookOpen,
  faDisplay,
  faFile,
  faCircleInfo,
  faEnvelope,
} from "@fortawesome/free-solid-svg-icons";
import { useSidebar } from "../../HOCs/SidebarProvider";
import { useSelector } from "react-redux";
import { RootState } from "../../redux/store";

const VendorSidebar = ({ sidebarOpen, closeSidebar }) => {
  const { sectionsVisible, toggleSection } = useSidebar();
  const userId = useSelector((state: RootState) => state.user.userId);
  const userAddons = useSelector((state: RootState) => state.user.addons);
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
          className={`vendor-sidebar__link  ${
            location.pathname === "/app/business/menu" ? "vendor-active" : ""
          }`}
          onClick={() => navigate("/app/business/menu")}
        >
          <FontAwesomeIcon icon={faBookOpen} className="vendor-sidebar__icon" />
          <p id="link">Menu</p>
        </div>

        {/* Appearance Section */}
        <div className="vendor-sidebar__section">
          <h2 onClick={() => toggleSection("appearance")}>
            <div>
              <FontAwesomeIcon
                icon={faDisplay}
                className="vendor-sidebar__icon"
              />
              Site Settings
            </div>
            {/* {sectionsVisible["appearance"] ? (
              <FontAwesomeIcon
                icon={faChevronUp}
                className="vendor-sidebar__icon"
              />
            ) : (
              <FontAwesomeIcon
                icon={faChevronDown}
                className="vendor-sidebar__icon"
              />
            )} */}
          </h2>
          {sectionsVisible["appearance"] && (
            <>
              <div
                className={`vendor-sidebar__link admin__link ${
                  location.pathname === "/manage/design" ? "vendor-active" : ""
                }`}
                onClick={() => navigate("/manage/design")}
              >
                <FontAwesomeIcon
                  icon={faUserPen}
                  className="vendor-sidebar__icon"
                />
                <p id="link">Information</p>
              </div>

              <div
                className={`vendor-sidebar__link admin__link ${
                  location.pathname === "/app/business/qr-codes"
                    ? "vendor-active"
                    : ""
                }`}
                onClick={() => navigate("/app/business/qr-codes")}
              >
                <FontAwesomeIcon
                  icon={faQrcode}
                  className="vendor-sidebar__icon"
                />
                <p id="link">Qr Codes</p>
              </div>

              <div
                className={`vendor-sidebar__link admin__link ${
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
                <p id="link">Appearance</p>
              </div>

              {userAddons && (userAddons?.landing || userAddons?.about) && (
                <div
                  className="vendor-sidebar__section"
                  style={{
                    marginLeft: "20px",
                  }}
                >
                  <h2 onClick={() => toggleSection("pages")}>
                    <div>
                      <FontAwesomeIcon
                        icon={faFile}
                        className="vendor-sidebar__icon"
                      />
                      Pages
                    </div>
                  </h2>
                  {sectionsVisible["pages"] && (
                    <>
                      {userAddons && userAddons?.landing && (
                        <div
                          className={`vendor-sidebar__link admin__link ${
                            location.pathname === "/manage/design"
                              ? "vendor-active"
                              : ""
                          }`}
                          onClick={() => navigate("/manage/design")}
                        >
                          <FontAwesomeIcon
                            icon={faGlobe}
                            className="vendor-sidebar__icon"
                          />
                          <p id="link">Home page</p>
                        </div>
                      )}
                      {userAddons && userAddons?.about && (
                        <div
                          className={`vendor-sidebar__link admin__link ${
                            location.pathname === "/manage/design"
                              ? "vendor-active"
                              : ""
                          }`}
                          onClick={() => navigate("/manage/design")}
                        >
                          <FontAwesomeIcon
                            icon={faCircleInfo}
                            className="vendor-sidebar__icon"
                          />
                          <p id="link">About page</p>
                        </div>
                      )}
                      {userAddons && userAddons?.contact && (
                        <div
                          className={`vendor-sidebar__link admin__link ${
                            location.pathname === "/manage/design"
                              ? "vendor-active"
                              : ""
                          }`}
                          onClick={() => navigate("/manage/design")}
                        >
                          <FontAwesomeIcon
                            icon={faEnvelope}
                            className="vendor-sidebar__icon"
                          />
                          <p id="link">Contact page</p>
                        </div>
                      )}
                    </>
                  )}
                </div>
              )}
            </>
          )}
        </div>
        {/* Appearance Section End */}

        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/business/menus" ? "vendor-active" : ""
          }`}
          onClick={() => navigate("/app/business/menus")}
        >
          <FontAwesomeIcon icon={faFolder} className="vendor-sidebar__icon" />
          <p id="link">My Menus</p>
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
          <p id="link">Billing</p>
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

export default React.memo(VendorSidebar);
