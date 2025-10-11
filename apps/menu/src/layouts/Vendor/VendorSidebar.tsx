import React from "react";
import { useNavigate } from "react-router-dom";
import {
  FaTimes,
  FaHome,
  FaChevronDown,
  FaChevronUp,
  FaCog,
  FaExternalLinkAlt,
  FaQrcode,
  FaMagic,
  FaStar,
  FaUserEdit,
  FaCreditCard,
  FaPuzzlePiece,
  FaFolder,
  FaChartLine,
  FaGlobe,
  FaBookOpen,
  FaDesktop,
  FaFile,
  FaInfoCircle,
  FaEnvelope,
  FaBullhorn,
  FaUserLock,
  FaRss,
  FaWallet,
} from "react-icons/fa";
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
        <FaTimes
          className="vendor-close__icon"
          onClick={() => closeSidebar()}
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
          <FaHome className="vendor-sidebar__icon" />
          <p id="link">Home</p>
        </div>
        <div
          className={`vendor-sidebar__link  ${
            location.pathname === "/app/business/menu" ? "vendor-active" : ""
          }`}
          onClick={() => navigate("/app/business/menu")}
        >
          <FaBookOpen className="vendor-sidebar__icon" />
          <p id="link">Menu</p>
        </div>

        {/* Appearance Section */}
        <div className="vendor-sidebar__section">
          <h2 onClick={() => toggleSection("appearance")}>
            <div>
              <FaDesktop
                className="vendor-sidebar__icon"
              />
              Site Settings
            </div>
            {/* {sectionsVisible["appearance"] ? (
              <FaChevronUp
                className="vendor-sidebar__icon"
              />
            ) : (
              <FaChevronDown
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
                <FaUserEdit
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
                <FaQrcode
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
                <FaMagic
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
                      <FaFile
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
                          <FaGlobe
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
                          <FaInfoCircle
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
                          <FaEnvelope
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
          <FaFolder className="vendor-sidebar__icon" />
          <p id="link">My Menus</p>
        </div>

        {userAddons && userAddons?.feedback && (
          <div
            className={`vendor-sidebar__link ${
              location.pathname === "/manage/design" ? "vendor-active" : ""
            }`}
            onClick={() => navigate("/manage/design")}
          >
            <FaStar className="vendor-sidebar__icon" />
            <p id="link">Feedback</p>
          </div>
        )}

        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/manage/design" ? "vendor-active" : ""
          }`}
          onClick={() => navigate("/manage/design")}
        >
          <FaChartLine
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
              <FaBullhorn
                className="vendor-sidebar__icon"
              />
              Marketing
            </div>
            {managementVisible ? (
              <FaChevronUp
                className="vendor-sidebar__icon"
              />
            ) : (
              <FaChevronDown
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
                <FaUserLock
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
                <FaRss
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
          <FaWallet className="vendor-sidebar__icon" />
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
          <FaCreditCard
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
          <FaCog className="vendor-sidebar__icon" />
          <p id="link">Settings</p>
        </div>
        <div
          className={`vendor-sidebar__link ${
            location.pathname === "/app/business/addons" ? "vendor-active" : ""
          }`}
          onClick={() => navigate("/app/business/addons")}
        >
          <FaPuzzlePiece
            className="vendor-sidebar__icon"
          />
          <p id="link"> Add-Ons</p>
        </div>
        <div
          className={`vendor-sidebar__link`}
          onClick={() => window.open(`/restaurant/${userId}`, "_blank")}
        >
          <FaExternalLinkAlt
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

export default React.memo(VendorSidebar);
