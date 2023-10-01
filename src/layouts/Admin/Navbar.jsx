import { Link, useNavigate } from "react-router-dom";
import { auth, db } from "../../firebase";
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

const Navbar = ({ openSidebar }) => {
  const navigate = useNavigate();
  const [notificationsCount, setNotificationsCount] = useState(0);

  useEffect(() => {
    if (auth.currentUser) {
      db.collection("contacts")
        .orderBy("createdAt", "desc")
        .get()
        .then((snapshot) => {
          setNotificationsCount(snapshot.size);
          console.log(snapshot.size);
        });
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="navbar">
      <div className="nav_icon" onClick={() => openSidebar("navbar")}>
        <FontAwesomeIcon icon={faBars} />
      </div>

      <div className="navbar__left">
        <p className="active_link">{auth.currentUser?.displayName}</p>
      </div>
      <div className="navbar__right">
        <Link to="/">
          <FontAwesomeIcon icon={faHome} size="xl" />
        </Link>
        <Link to="/manage/notifications">
          <Badge
            badgeContent={notificationsCount}
            color="error"
            anchorOrigin={{
              vertical: "top",
              horizontal: "right",
            }}
          >
            <FontAwesomeIcon icon={faEnvelope} size="xl" />
          </Badge>
        </Link>
        <div
          to="/auth/profile"
          onClick={() => {
            auth.signOut().then(() => {
              navigate("/");
            });
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
