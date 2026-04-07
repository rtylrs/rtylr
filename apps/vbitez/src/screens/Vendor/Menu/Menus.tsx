import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useMenuStore } from "../../../stores/menuStore";
import { useUserStore } from "../../../stores/userStore";

const Menus = () => {
  const navigate = useNavigate();
  const { userId } = useUserStore();
  const { allMenus, activeMenuId, status, error, fetchAllMenus, setActiveMenu, createNewMenu } = useMenuStore();
  const [newMenuName, setNewMenuName] = useState("");
  const [showCreateMenuForm, setShowCreateMenuForm] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");

  useEffect(() => {
    fetchAllMenus(userId as string);
  }, [fetchAllMenus, userId]);

  const handleSetActiveMenu = (menuId) => {
    setActiveMenu(userId as string, menuId)
      .then(() => alert("Active menu updated successfully"))
      .catch((err) => console.error("Error setting active menu:", err));
  };

  const handleViewMenu = (menuId) => {
    navigate(`/app/business/menu/${menuId}`);
  };

  const handleCreateNewMenu = () => {
    if (newMenuName.trim()) {
      createNewMenu(userId as string, newMenuName)
        .then(() => {
          setNewMenuName("");
          setShowCreateMenuForm(false);
          alert("New menu created successfully");
        })
        .catch((err) => console.error("Error creating new menu:", err));
    } else {
      alert("Menu name cannot be empty");
    }
  };

  const handleSearchChange = (e) => {
    setSearchTerm(e.target.value);
  };

  const filteredMenus = allMenus.filter((menu: any) =>
    menu.menuName.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div>
      <h1>Menus</h1>
      <input
        type="text"
        value={searchTerm}
        onChange={handleSearchChange}
        placeholder="Search menus..."
        style={{ marginBottom: "20px", width: "100%", padding: "8px" }}
      />
      {status === "loading" && <p>Loading...</p>}
      {status === "failed" && <p>{error}</p>}
      {status === "succeeded" && (
        <>
          {Array.isArray(filteredMenus) && filteredMenus.length === 0 ? (
            <div>
              <p>No menus found. Create a new menu:</p>
              <input
                type="text"
                value={newMenuName}
                onChange={(e) => setNewMenuName(e.target.value)}
                placeholder="New menu name"
              />
              <button onClick={handleCreateNewMenu}>Create New Menu</button>
            </div>
          ) : (
            <>
              <ul>
                {Array.isArray(filteredMenus) &&
                  filteredMenus.map((menu: any) => (
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
                {showCreateMenuForm ? (
                  <div>
                    <input
                      type="text"
                      value={newMenuName}
                      onChange={(e) => setNewMenuName(e.target.value)}
                      placeholder="New menu name"
                    />
                    <button onClick={handleCreateNewMenu}>
                      Create New Menu
                    </button>
                  </div>
                ) : (
                  <button onClick={() => setShowCreateMenuForm(true)}>
                    Create New Menu
                  </button>
                )}
              </div>
            </>
          )}
        </>
      )}
    </div>
  );
};

export default Menus;
