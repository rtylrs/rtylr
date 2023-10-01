import React, { useState } from "react";
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
  faPlusCircle,
  faChevronDown,
  faChevronUp,
  faShoppingBag,
  faCog,
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
    petitionsVisible,
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
        <div className="sidebar__link ">
          <FontAwesomeIcon icon={faHome} className="sidebar__icon" />
          <Link to="/manage" id="link">
            My Dashboard
          </Link>
        </div>
        <div className="sidebar__link">
          <FontAwesomeIcon icon={faShoppingBag} className="sidebar__icon" />
          <Link to="/manage/products" id="link">
            Sanity Products
          </Link>
        </div>
        <div
          className="sidebar__section"
          onClick={() => toggleSection("customers")}
        >
          <h2>
            Customers
            {customersVisible ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
            )}
          </h2>
          {customersVisible && (
            <>
              <div className="sidebar__link">
                <FontAwesomeIcon icon={faUser} className="sidebar__icon" />
                <Link to="/manage/users" id="link">
                  Users
                </Link>
              </div>
              <div className="sidebar__link">
                <FontAwesomeIcon icon={faBuilding} className="sidebar__icon" />
                <Link to="/manage/vendors" id="link">
                  Suppliers
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
            Management
            {managementVisible ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
            )}
          </h2>
          {managementVisible && (
            <>
              <div className="sidebar__link">
                <FontAwesomeIcon
                  icon={faUserSecret}
                  className="sidebar__icon"
                />
                <Link to="/manage/managment" id="link">
                  Add/Remove Admins
                </Link>
              </div>
              <div className="sidebar__link">
                <FontAwesomeIcon icon={faUser} className="sidebar__icon" />
                <Link to="/manage/feedback" id="link">
                  Feedback
                </Link>
              </div>
              <div className="sidebar__link">
                <FontAwesomeIcon icon={faCube} className="sidebar__icon" />
                <Link to="/manage/blog" id="link">
                  Blog
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
            Configuration
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
            Blog
            {blogVisible ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
            )}
          </h2>
          {blogVisible && (
            <>
              <div className="sidebar__link">
                <FontAwesomeIcon icon={faCube} className="sidebar__icon" />
                <Link to="/manage/blog" id="link">
                  Blog Section
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
            Reports
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
            Orders
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
          onClick={() => toggleSection("petitions")}
        >
          <h2>
            Petitions
            {petitionsVisible ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
            )}
          </h2>
          {petitionsVisible && (
            <>
              <div className="sidebar__link">
                <FontAwesomeIcon
                  icon={faPlusCircle}
                  className="sidebar__icon"
                />
                <Link to="/manage/petitions" id="link">
                  Add Petition
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
            Donations
            {donationsVisible ? (
              <FontAwesomeIcon icon={faChevronUp} className="sidebar__icon" />
            ) : (
              <FontAwesomeIcon icon={faChevronDown} className="sidebar__icon" />
            )}
          </h2>
          {donationsVisible && <>{/* Add donation links here */}</>}
        </div>

        <div className="sidebar__link">
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
