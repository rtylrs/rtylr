import { faHeartBroken } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

const AdminAddons = () => {
  return (
    <div className="admin-addons">
      <FontAwesomeIcon icon={faHeartBroken} />
      <h1>Apologies</h1>
      <p>
        We apologize, but Add-Ons are currently not available in the{" "}
        <strong>starter</strong> package. If you would like to upgrade your
        package and access premium Add-Ons and other features, please contact
        support for more information.
      </p>
    </div>
  );
};

export default AdminAddons;
