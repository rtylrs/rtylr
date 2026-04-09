import { Spinner } from '@voxire/components';
import { useState } from "react";
import { Star, Info, Phone, Globe } from "phosphor-react";
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
    <div className="relative space-y-6">
      {planType === "free" && (
        <div className="absolute inset-0 bg-white/80 backdrop-blur-sm z-10 flex flex-col items-center justify-center gap-4 rounded-xl">
          <div className="text-base font-semibold text-primary">
            Addons are not available in the free plan
          </div>
          <Button onClick={() => navigate("/app/business/billing")}>
            Upgrade
          </Button>
        </div>
      )}
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-primary">Add-Ons</h1>
        <div className="flex items-center gap-3">
          <Button
            onClick={() => setShowInstalled(!showInstalled)}
            styles={{ height: "35px" }}
          >
            {showInstalled ? "Show All" : "Installed"}
          </Button>
          <input
            type="text"
            placeholder="Search for an add-on"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="border border-black/10 rounded-lg px-3 py-1.5 text-sm bg-white"
          />
        </div>
      </div>
      <div className="space-y-6">
        <div>
          <h2 className="text-base font-semibold text-primary mb-4">Power-Ups</h2>
          <div className="grid grid-cols-3 gap-4">
            {filteredAddons.map((addon) => (
              <div key={addon.key} className="bg-white rounded-xl p-5 border border-black/5 shadow-card flex flex-col gap-3">
                <addon.icon size={24} className="text-accent" />
                <h3 className="font-semibold text-primary">{addon.name}</h3>
                <p className="text-sm text-muted">{addon.description}</p>
                {keyLoading === (addon.key as string | null) ? (
                  <Spinner />
                ) : (
                  <button
                    className={`px-3 py-1.5 text-sm rounded-lg font-medium transition-colors ${
                      userAddons?.[addon.key]
                        ? 'bg-green-100 text-green-700 hover:bg-green-200'
                        : 'bg-accent text-white hover:opacity-90'
                    }`}
                    onClick={() => handleToggleAddon(addon.key)}
                  >
                    {userAddons?.[addon.key] ? "Remove" : "Install (Free)"}
                  </button>
                )}
              </div>
            ))}
          </div>
        </div>
        <div className="text-center py-8 text-muted text-sm">
          <h3>More Addons Coming Soon!</h3>
        </div>
      </div>
    </div>
  );
};

export default Addons;
