import { Spinner } from '@voxire/components';
import { useState } from "react";
import { Star, Info, Phone, Globe } from "lucide-react";
import useAsync from "../../hooks/useAsync";
import { userApi } from "../../api";
import { Button } from "../../components/Button";
import { useNavigate } from "react-router-dom";
import { useUserStore } from "../../stores/userStore";
import notify from "../../lib/notify/notify";

const ADDONS = [
  {
    key: "feedback",
    name: "Feedback",
    description: "Collect feedback from your customers.",
    icon: Star,
  },
  {
    key: "about",
    name: "About",
    description: "Information about your business.",
    icon: Info,
  },
  {
    key: "contact",
    name: "Contact",
    description: "Enable customers to contact you easily.",
    icon: Phone,
  },
  {
    key: "landing",
    name: "Landing",
    description: "Create a landing page for your business.",
    icon: Globe,
  },
];

const Addons = () => {
  const { user, addons: userAddons, userId, getUserInfo } = useUserStore();
  const planType = user?.planType;
  const [searchTerm, setSearchTerm] = useState("");
  const [showInstalled, setShowInstalled] = useState(false);
  const navigate = useNavigate();
  const [keyLoading, setKeyLoading] = useState<number | null>();

  const editAddonsFunc = useAsync({
    fn: userApi.editAddons,
    onSuccess: () => {
      getUserInfo().then(() => {
        notify.success("Updated Successfully");
        setKeyLoading(null);
      });
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

      setKeyLoading(addonKey);
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
          <Button onClick={() => navigate("/app/business/billing")}>
            Upgrade
          </Button>
        </div>
      )}
      <div className="addons-header">
        <h1>Add-Ons</h1>
        <div>
          <Button
            onClick={() => setShowInstalled(!showInstalled)}
            styles={{
              height: "35px",
            }}
          >
            {showInstalled ? "Show All" : "Installed"}
          </Button>
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
                <addon.icon size={24} className="addon-icon" />
                <h3>{addon.name}</h3>
                <p>{addon.description}</p>
                {keyLoading === (addon.key as string | null) ? (
                  <Spinner />
                ) : (
                  <button
                    className={userAddons?.[addon.key] ? "installed" : ""}
                    onClick={() => handleToggleAddon(addon.key)}
                  >
                    {userAddons?.[addon.key] ? "Remove" : "Install (Free)"}
                  </button>
                )}
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
