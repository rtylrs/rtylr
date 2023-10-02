import React from "react";

function Purchase() {
  return (
    <div>
      <div className={"deliveryAndPickupOptions"}>
        <button className={"delivery"}>Delivery</button>
        <button className={"pickup"}>Pickup</button>
        <button className={"hotel"}>Hotel</button>
      </div>
    </div>
  );
}

export default Purchase;
