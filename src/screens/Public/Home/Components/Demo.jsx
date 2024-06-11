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
    <h2>{t("Demo.title")}</h2>
    <p>{t("Demo.description")}</p>
    <div className="qr-code">
      <img src="/images/VoxireWebsiteQR.png" alt="QR Code" />
    </div>
    <VbitezButton variant="abed">{t("Demo.button")}</VbitezButton>
  </div>


)

}

export default Demo;