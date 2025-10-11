import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { FaPen, FaEllipsisV } from "react-icons/fa";

function MenuItem({ item }) {
  const navigate = useNavigate();
  const [showMenu, setShowMenu] = useState(false);

  const handleMenuClick = () => {
    setShowMenu(!showMenu);
  };

  const handleMenuClose = () => {
    setShowMenu(false);
  };

  const handleDelete = () => {
    handleMenuClose();
  };

  const handleDuplicate = () => {
    handleMenuClose();
  };

  const handleHide = () => {
    handleMenuClose();
  };

  return (
    <div className="menu-item">
      <img src={item.coverImage} alt={item.name} className="menu-item-image" />
      <div className="item-details">
        <h3>{item.name}</h3>
        <p>Price: {item.price} €</p>
        <p>Description: {item.description}</p>
      </div>
      <div className="item-controls">
        <button
          className="edit-button"
          onClick={() => {
            navigate(`/app/business/edit-item/${item._id}`);
          }}
        >
          <FaPen />
        </button>
        <button className="menu-button" onClick={handleMenuClick}>
          <FaEllipsisV />
        </button>
        {showMenu && (
          <div className="dropdown-menu">
            <button onClick={handleDuplicate}>Duplicate item</button>
            <button onClick={handleDelete}>Delete item</button>
            <button onClick={handleHide}>Hide item</button>
          </div>
        )}
      </div>
    </div>
  );
}

export default MenuItem;
