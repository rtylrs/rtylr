import { Link, useNavigate } from "react-router-dom";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faBars,
  faEnvelope,
  faHome,
  faRightFromBracket,
} from "@fortawesome/free-solid-svg-icons";
import PropTypes from "prop-types";
import { Badge } from "@mui/material";
import { useState, useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { logout } from "../../redux/reducers/userSlice";

const Navbar = ({ openSidebar }) => {
  const navigate = useNavigate();
  const [notificationsCount, setNotificationsCount] = useState(0);
  const dispatch = useDispatch();
  const user = useSelector((state) => state.user);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="navbar">
      <div className="nav_icon" onClick={() => openSidebar("navbar")}>
        <FontAwesomeIcon icon={faBars} />
      </div>

      <div className="navbar__left">
        <p className="active_link">{user?.user?.name}</p>
      </div>
      <div className="navbar__right">
        <Link to="/manage/contacts">
          <Badge
            badgeContent={notificationsCount}
            color="error"
            anchorOrigin={{
              vertical: "top",
              horizontal: "right",
            }}
            className="badgeAdmin"
          >
            <FontAwesomeIcon icon={faEnvelope} size="xl" />
          </Badge>
        </Link>
        <div
          to="/auth/login"
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

Navbar.propTypes = {
  openSidebar: PropTypes.func.isRequired,
};

export default Navbar;
