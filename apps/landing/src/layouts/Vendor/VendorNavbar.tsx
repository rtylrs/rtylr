import React from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faBars } from "@fortawesome/free-solid-svg-icons";
import { useDispatch, useSelector } from "react-redux";
import { logout } from "../../redux/reducers/userSlice";
import { RootState } from "../../redux/store";

const VendorNavbar: React.FC<{
  openSidebar: (sidebarOpen: boolean) => void;
  sidebarOpen: boolean;
}> = ({ openSidebar }) => {
  const dispatch = useDispatch();
  const user = useSelector((state: RootState) => state.user);

  return (
    <div className="vendor-navbar">
      <div className="vendor_nav_icon" onClick={() => openSidebar(true)}>
        <FontAwesomeIcon icon={faBars} />
      </div>
      <div className="vendor-navbar__left">
        <p className="active_link">
          {user?.user?.name}{" "}
          <span className="plan-type">({user?.user?.planType})</span>
        </p>
      </div>
      <div className="vendor-navbar__right">
        <div
          onClick={() => {
            dispatch(logout());
          }}
        >
          <p>Log out</p>
        </div>
      </div>
    </div>
  );
};

export default VendorNavbar;
