import NewsAndUpdates from "./Components/NewsAndUpdates";
import QuickLinks from "./Components/QuickLinks";

const VendorDashboard = (): JSX.Element => {
  return (
    <div className="vendor-dashboard">
      <div className="main-content">
        <h3>Quick Links</h3>
        <QuickLinks />
      </div>
      <div className="news-updates">
        <h3>News and Updates</h3>
        <NewsAndUpdates />
      </div>
    </div>
  );
};

export default VendorDashboard;
