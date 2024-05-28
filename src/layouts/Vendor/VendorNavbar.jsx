import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faBars, faRightFromBracket } from "@fortawesome/free-solid-svg-icons";
import PropTypes from "prop-types";
import { useDispatch, useSelector } from "react-redux";
import { logout } from "../../redux/reducers/userSlice";

const VendorNavbar = ({ openSidebar }) => {
  const dispatch = useDispatch();
  const user = useSelector((state) => state.user);

  return (
    <div className="vendor-navbar">
      <div className="vendor_nav_icon" onClick={() => openSidebar("navbar")}>
        <FontAwesomeIcon icon={faBars} />
      </div>
      <div className="vendor-navbar__left">
        <p className="active_link">{user?.user?.name}</p>
      </div>
      <div className="vendor-navbar__right">
        <div
          to="/auth/profile"
          onClick={() => {
            dispatch(logout());
          }}
        >
          <FontAwesomeIcon icon={faRightFromBracket} size="xl" />
        </div>
      </div>
    </div>
  );
};

VendorNavbar.propTypes = {
  openSidebar: PropTypes.func.isRequired,
};

export default VendorNavbar;
