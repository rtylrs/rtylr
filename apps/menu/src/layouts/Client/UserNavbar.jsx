import { Link, useNavigate } from "react-router-dom";
import { auth } from "../../firebase";
import {
  FaBars,
  FaHome,
  FaSignOutAlt,
} from "react-icons/fa";

const Navbar = ({ openSidebar }) => {
  const navigate = useNavigate();
  return (
    <div className="navbar">
      <div className="nav_icon" onClick={() => openSidebar("navbar")}>
        <FaBars />
      </div>

      <div className="navbar__left">
        <p className="active_link">{auth.currentUser?.displayName}</p>
      </div>
      <div className="navbar__right">
        <Link to="/">
          <FaHome size="xl" />
        </Link>
        <div
          to="/auth/profile"
          onClick={() => {
            auth.signOut().then(() => {
              navigate("/");
            });
          }}
        >
          <FaSignOutAlt size="xl" />
        </div>
      </div>
    </div>
  );
};

export default Navbar;
