import Checks from "./Components/Checks";
import NewsAndUpdates from "./Components/NewsAndUpdates";
import QuickLinks from "./Components/QuickLinks";
import Visits from "./Components/Visits";

const VendorDashboard = (): JSX.Element => {
  return (
    <div className="vendor-dashboard">
      <div className="main-content">
        <h3>Quick Links</h3>
        <QuickLinks />
        <h3
          style={{
            marginTop: "20px",
          }}
        >
          Checks
        </h3>
        <Checks />
        <h3
          style={{
            marginTop: "20px",
          }}
        >
          Visits
        </h3>
        <Visits />
      </div>
      <div className="news-updates">
        <h3>News and Updates</h3>
        <NewsAndUpdates />
      </div>
    </div>
  );
};

export default VendorDashboard;
