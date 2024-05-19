import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import VbitezButton from "../../../../components/Buttons/Buttons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import useColors from "../../../../hooks/useColors";


function Demo(){
    const { t } = useTranslation();
  const navigate = useNavigate();
  const { accent } = useColors();
  return(

    <div className="demo-menu-container">
    <h2>See a demo online menu</h2>
    <p>Just open the camera on your phone and scan the <span>QR code</span> below!</p>
    <div className="qr-code">
      <img src="/images/VoxireWebsiteQR.png" alt="QR Code" />
    </div>
    <VbitezButton variant="abed">Create a menu for you, now!</VbitezButton>
  </div>


)

}

export default Demo;