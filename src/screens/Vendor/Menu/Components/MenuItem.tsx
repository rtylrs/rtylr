import React, { useState } from "react";
import PropTypes from "prop-types";
import { useNavigate } from "react-router-dom";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faPen, faEllipsisV } from "@fortawesome/free-solid-svg-icons";

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
    console.log(item);
    handleMenuClose();
  };

  const handleDuplicate = () => {
    console.log("Duplicate item");
    handleMenuClose();
  };

  const handleHide = () => {
    console.log("Hide item");
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
          <FontAwesomeIcon icon={faPen} />
        </button>
        <button className="menu-button" onClick={handleMenuClick}>
          <FontAwesomeIcon icon={faEllipsisV} />
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

MenuItem.propTypes = {
  item: PropTypes.object.isRequired,
};

export default MenuItem;
