import React from "react";
import { Link } from "react-router-dom";

function QrCodes() {
  return (
    <div>
      <Link to="/app/business/qr-codes/create">Create Qr</Link>
    </div>
  );
}

export default QrCodes;
