import { Link, useNavigate } from "react-router-dom";
import { auth } from "../../firebase";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faBars, faHome, faExplosion } from "@fortawesome/free-solid-svg-icons";
import PropTypes from "prop-types";

const Navbar = ({ openSidebar }) => {
  const navigate = useNavigate();
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
        <div
          to="/auth/profile"
          onClick={() => {
            auth.signOut().then(() => {
              navigate("/");
            });
          }}
        >
          <FontAwesomeIcon icon={faExplosion} size="xl" />
        </div>
      </div>
    </div>
  );
};

Navbar.propTypes = {
  openSidebar: PropTypes.func.isRequired,
};

export default Navbar;
