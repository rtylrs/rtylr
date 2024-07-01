import PropTypes from "prop-types";
import { useNavigate } from "react-router-dom";

function MenuItem({ item }) {
  const navigate = useNavigate();

  const handleDelete = () => {
    console.log(item);
  };

  return (
    <div className="menu-item">
      <img src={item.coverImage} alt={item.name} height={50} width={50} />
      <div className="item-details">
        <h3>{item.name}</h3>
        <p>Options: {item.options}</p>
        <div className="item-controls">
          <label>Visible</label>
          <label>Popular</label>
          <label>In Stock</label>
          <p>Price: {item.price} €</p>
          <p>Price: {item.description} €</p>
          <button
            onClick={() => {
              navigate(`/app/business/edit-item/${item._id}`);
            }}
          >
            Edit
          </button>
          <button onClick={() => handleDelete()}>Delete</button>
        </div>
      </div>
    </div>
  );
}

export default MenuItem;

MenuItem.propTypes = {
  item: PropTypes.object.isRequired,
};
