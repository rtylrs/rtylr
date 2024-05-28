import React from "react";
import { Link } from "react-router-dom";

function QrCodes() {
  return (
    <div>
      <Link to="/app/restaurant/qr-codes/create">Create Qr</Link>
    </div>
  );
}

export default QrCodes;
