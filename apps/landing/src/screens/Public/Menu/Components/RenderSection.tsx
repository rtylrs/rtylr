import MenuItem from "./MenuItem";

const RenderSection = (item) => {
  if (item.type === "category") {
    return (
      <div key={item._id} className="category-section" id={item._id}>
        <div className="category-header">
          <div>
            <span>{item?.details?.name}</span>
            {item?.details?.icon && (
              <img src={item?.details?.icon} alt={item?.details?.name} />
            )}
          </div>
        </div>
        <div className="category-items">
          {item?.details?.menuItems?.map((menuItem) => (
            <MenuItem key={menuItem._id} item={menuItem} />
          ))}
        </div>
      </div>
    );
  } else if (item.type === "banner") {
    return (
      <div key={item._id} className="banner-section">
        <p>{item?.details?.banner?.title}</p>
        <img
          src={item?.details?.banner?.bannerImage}
          alt={item?.details?.title}
          style={{
            width: "100%",
            height: "100%",
            objectFit: "cover",
            objectPosition: "center",
          }}
        />
      </div>
    );
  }
  return null;
};

export default RenderSection;
