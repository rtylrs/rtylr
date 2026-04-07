import { invoke } from "@tauri-apps/api/tauri";
import Checks from "./Components/Checks";
import NewsAndUpdates from "./Components/NewsAndUpdates";
import QuickLinks from "./Components/QuickLinks";
import Visits from "./Components/Visits";

const VendorDashboard = (): JSX.Element => {
  const handleOpenCashDrawer = async () => {
    try {
      const result = await invoke("open_cash_drawer", {
        printerName: "POSPrinter POS-80C",
      });
      console.log(result);
      alert("Cash drawer opened successfully!");
    } catch (error) {
      console.error("Failed to open cash drawer:", error);
      alert(`Failed to open cash drawer: ${error}`);
    }
  };

  return (
    <div className="vendor-dashboard">
      <div className="main-content">
        <h3>Quick Links</h3>
        <QuickLinks />
        <div style={{ marginTop: "20px" }}>
          <button
            onClick={handleOpenCashDrawer}
            style={{
              padding: "10px 20px",
              backgroundColor: "#4CAF50",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
              fontSize: "16px",
            }}
          >
            Open Cash Drawer
          </button>
        </div>
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
          Daily Visits
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
