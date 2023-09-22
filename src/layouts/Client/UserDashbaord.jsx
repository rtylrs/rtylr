import PropTypes from "prop-types";
import { auth } from "../../firebase";
import { useDispatch } from "react-redux";
import { Link, useNavigate } from "react-router-dom";
import { logout } from "../../redux/reducers/userSlice";

import Logo from "../../assets/logo.png";

function DashboardLayout({ children }) {
  const dispatch = useDispatch();
  const navigate = useNavigate();

  return (
    <div className="dashboard-layout">
      <nav className="sidebar">
        <div className="logo">
          <Link to="/">
            <img src={Logo} alt="logo" />
          </Link>
        </div>
        <ul>
          <li>
            <Link to="#">Profile</Link>
          </li>
          <li>
            <Link to="#">Favorites</Link>
          </li>
          <li>
            <Link to="#">Previous Orders</Link>
          </li>
          <li>
            <Link to="#">Become a Business</Link>
          </li>
        </ul>
      </nav>

      <main className="main-content">
        <header className="top-bar">
          <p>Hello {auth.currentUser.displayName}!</p>
          <button
            className="logout-button"
            onClick={() => {
              auth.signOut().then(() => {
                dispatch(logout());
                navigate("/");
              });
            }}
          >
            Logout
          </button>
        </header>

        {children}
      </main>
    </div>
  );
}

DashboardLayout.propTypes = {
  children: PropTypes.node.isRequired,
};

export default DashboardLayout;
