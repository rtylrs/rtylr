import PropTypes from "prop-types";

function MenuItem({ item }) {
  return (
    <div className="menu-item">
      <img src={item.coverImage} alt={item.name} height={50} width={50} />
      <div className="item-details">
        <h3>{item.name}</h3>
        <div className="item-controls">
          <p>Price: {item.price} $</p>
        </div>
      </div>
    </div>
  );
}

export default MenuItem;

MenuItem.propTypes = {
  item: PropTypes.object.isRequired,
  onEdit: PropTypes.func.isRequired,
};
