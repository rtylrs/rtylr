import PropTypes from "prop-types";
import { auth } from "../../firebase";
import { useDispatch } from "react-redux";
import { Link } from "react-router-dom";

function DashboardLayout({ children }) {
  return (
    <div className="dashboard-layout">
      <nav className="sidebar">
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
          <p>{auth.currentUser.displayName}</p>
          <button className="logout-button">Logout</button>
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
