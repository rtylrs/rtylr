import PropTypes from "prop-types";

function MenuItem({ item, onEdit }) {
  return (
    <div className="menu-item">
      {console.log(item)}
      <img src={item.coverImage} alt={item.name} height={50} width={50} />
      <div className="item-details">
        <h3>{item.name}</h3>
        <p>Options: {item.options}</p>
        <div className="item-controls">
          <label>
            <input
              type="checkbox"
              checked={item.visible}
              onChange={() => onEdit({ ...item, visible: !item.visible })}
            />
            Visible
          </label>
          <label>
            <input
              type="checkbox"
              checked={item.popular}
              onChange={() => onEdit({ ...item, popular: !item.popular })}
            />
            Popular
          </label>
          <label>
            <input
              type="checkbox"
              checked={item.inStock}
              onChange={() => onEdit({ ...item, inStock: !item.inStock })}
            />
            In Stock
          </label>
          <p>Price: {item.price} €</p>
          <button onClick={() => onEdit(item)}>Edit</button>
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
