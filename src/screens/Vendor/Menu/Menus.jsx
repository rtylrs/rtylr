import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import {
  fetchAllMenus,
  setActiveMenuId,
  createNewMenu,
} from "../../../redux/reducers/menuSlice";
import { useNavigate } from "react-router-dom";

const Menus = () => {
  const navigate = useNavigate();
  const userId = useSelector((state) => state.user.userId);
  const dispatch = useDispatch();
  const { allMenus, activeMenuId, status, error } = useSelector(
    (state) => state.menu
  );
  const [newMenuName, setNewMenuName] = useState("");

  useEffect(() => {
    dispatch(fetchAllMenus({ restaurantId: userId }));
  }, [dispatch, userId]);

  const handleSetActiveMenu = (menuId) => {
    dispatch(setActiveMenuId({ restaurantId: userId, menuId }))
      .then(() => alert("Active menu updated successfully"))
      .catch((err) => console.error("Error setting active menu:", err));
  };

  const handleViewMenu = (menuId) => {
    navigate(`/app/business/menu/${menuId}`);
  };

  const handleCreateNewMenu = () => {
    if (newMenuName.trim()) {
      dispatch(createNewMenu({ restaurantId: userId, menuName: newMenuName }))
        .then(() => {
          setNewMenuName("");
          alert("New menu created successfully");
        })
        .catch((err) => console.error("Error creating new menu:", err));
    } else {
      alert("Menu name cannot be empty");
    }
  };

  return (
    <div>
      <h1>Menus</h1>
      {status === "loading" && <p>Loading...</p>}
      {status === "failed" && <p>{error}</p>}
      {status === "succeeded" && (
        <>
          <ul>
            {allMenus &&
              allMenus.map((menu) => (
                <li key={menu._id} style={{ marginBottom: "20px" }}>
                  <h2>{menu.menuName}</h2>
                  <button onClick={() => handleViewMenu(menu._id)}>
                    View Menu
                  </button>
                  <button
                    onClick={() => handleSetActiveMenu(menu._id)}
                    disabled={menu._id === activeMenuId}
                    style={{ marginLeft: "10px" }}
                  >
                    {menu._id === activeMenuId
                      ? "Active Menu"
                      : "Set as Active Menu"}
                  </button>
                </li>
              ))}
          </ul>
          <div>
            <input
              type="text"
              value={newMenuName}
              onChange={(e) => setNewMenuName(e.target.value)}
              placeholder="New menu name"
            />
            <button onClick={handleCreateNewMenu}>Create New Menu</button>
          </div>
        </>
      )}
    </div>
  );
};

export default Menus;
