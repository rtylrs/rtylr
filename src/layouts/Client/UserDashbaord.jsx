import PropTypes from "prop-types";

function DashboardLayout({ children }) {
  return (
    <div className="dashboard-layout">
      <nav className="sidebar">
        <ul>
          <li>
            <a href="#">Profile</a>
          </li>
          <li>
            <a href="#">Favorites</a>
          </li>
          <li>
            <a href="#">Previous Orders</a>
          </li>
          <li>
            <a href="#">Become a Business</a>
          </li>
        </ul>
      </nav>

      <main className="main-content">
        <header className="top-bar">
          <img src="user-image.jpg" alt="User Image" />
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
