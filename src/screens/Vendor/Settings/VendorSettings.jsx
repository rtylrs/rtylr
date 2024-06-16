const settingsOptions = [
  {
    title: "General",
    description: "View and update your store details",
    icon: "icon-general",
  },
  {
    title: "Plan and permissions",
    description:
      "View plan information and manage what staff can see or do in your store",
    icon: "icon-plan",
  },
  {
    title: "Payments",
    description: "Enable and manage your store's payment providers",
    icon: "icon-payments",
  },
  {
    title: "Files",
    description: "Upload images, videos, and documents",
    icon: "icon-files",
  },
  {
    title: "Backups",
    description: "Check your last backups",
    icon: "icon-backups",
  },
  {
    title: "Store languages",
    description: "Manage the languages your customers can view on your store",
    icon: "icon-languages",
  },
  {
    title: "Legal",
    description: "Manage your store's legal pages",
    icon: "icon-legal",
  },
];

const VendorSettings = () => {
  return (
    <div className="vendor-settings-container">
      <h1>Settings</h1>
      <div className="settings-grid">
        {settingsOptions.map((option) => (
          <div key={option.title} className="settings-item">
            <div className={`settings-icon ${option.icon}`}></div>
            <div className="settings-content">
              <h2>{option.title}</h2>
              <p>{option.description}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default VendorSettings;
