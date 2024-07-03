import { useState } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faStar,
  faInfoCircle,
  faPhone,
  faGlobe,
} from "@fortawesome/free-solid-svg-icons";
import { useDispatch, useSelector } from "react-redux";
import { AppDispatch, RootState } from "../../redux/store";
import useAsync from "../../hooks/useAsync";
import userApi from "../../api/user";
import VbitezButton from "../../components/Buttons/Buttons";
import { useNavigate } from "react-router-dom";
import { getUserInfo } from "../../redux/reducers/userSlice";
import notify from "../../lib/notify/notify";

const ADDONS = [
  {
    key: "feedback",
    name: "Feedback",
    description: "Collect feedback from your customers.",
    icon: faStar,
  },
  {
    key: "about",
    name: "About",
    description: "Information about your business.",
    icon: faInfoCircle,
  },
  {
    key: "contact",
    name: "Contact",
    description: "Enable customers to contact you easily.",
    icon: faPhone,
  },
  {
    key: "landing",
    name: "Landing",
    description: "Create a landing page for your business.",
    icon: faGlobe,
  },
];

const Addons = () => {
  const planType = useSelector((state: RootState) => state.user.user.planType);
  const userAddons = useSelector((state: RootState) => state.user.addons);
  const userId = useSelector((state: RootState) => state.user.userId);
  const [searchTerm, setSearchTerm] = useState("");
  const [showInstalled, setShowInstalled] = useState(false);
  const navigate = useNavigate();
  const dispatch = useDispatch<AppDispatch>();

  const editAddonsFunc = useAsync({
    fn: userApi.editAddons,
    onSuccess: () => {
      notify.success("Updated Successfully");
      dispatch(getUserInfo());
    },
    onError: () => {
      alert(editAddonsFunc.error);
    },
    callOnMount: false,
  });

  const handleToggleAddon = (addonKey) => {
    const newAddons = {
      ...userAddons,
      [addonKey]: !userAddons[addonKey],
    };

    if (planType !== "free") {
      editAddonsFunc.main({
        id: userId,
        addons: newAddons,
      });
    }
  };

  const filteredAddons = ADDONS.filter(
    (addon) =>
      addon.name.toLowerCase().includes(searchTerm.toLowerCase()) &&
      (!showInstalled || userAddons[addon.key])
  );

  return (
    <div className="addons">
      {planType === "free" && (
        <div className="overlay">
          <div className="overlay-message">
            Addons are not available in the free plan
          </div>
          <VbitezButton onClick={() => navigate("/app/business/billing")}>
            Upgrade
          </VbitezButton>
        </div>
      )}
      <div className="addons-header">
        <h1>Add-Ons</h1>
        <div>
          <VbitezButton
            onClick={() => setShowInstalled(!showInstalled)}
            styles={{
              height: "35px",
            }}
          >
            {showInstalled ? "Show All" : "Installed"}
          </VbitezButton>
          <input
            type="text"
            placeholder="Search for an add-on"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
      </div>
      <div className="addons-content">
        <div className="addon-category">
          <h2>Power-Ups</h2>
          <div className="addon-grid">
            {filteredAddons.map((addon) => (
              <div key={addon.key} className="addon-card">
                <FontAwesomeIcon icon={addon.icon} className="addon-icon" />
                <h3>{addon.name}</h3>
                <p>{addon.description}</p>
                <button
                  className={userAddons?.[addon.key] ? "installed" : ""}
                  onClick={() => handleToggleAddon(addon.key)}
                >
                  {userAddons?.[addon.key] ? "Remove" : "Install (Free)"}
                </button>
              </div>
            ))}
          </div>
        </div>
        <div className="more-addons">
          <h3>More Addons Coming Soon!</h3>
        </div>
      </div>
    </div>
  );
};

export default Addons;
