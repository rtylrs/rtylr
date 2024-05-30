import React from "react";

function MenuItem({ item, onEdit }) {
  return (
    <div>
      <h2>{item.name}</h2>
      <p>{item.description}</p>
      <p>{item.price}</p>
      <p>Categories: {item.categories.join(", ")}</p>
      <p>Allergens: {item.allergens.join(", ")}</p>
      <button onClick={onEdit}>Edit</button>
    </div>
  );
}

export default MenuItem;
