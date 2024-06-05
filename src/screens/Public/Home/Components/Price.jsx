import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import VbitezButton from "../../../../components/Buttons/Buttons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import useColors from "../../../../hooks/useColors";
import { faStar } from "@fortawesome/free-solid-svg-icons";

function Price() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { accent } = useColors();

  return (
    <div className="Prices">
      <h1 className="vbitez-title price-text">
        Choice the right pricing plan <br />
        that suits your need
      </h1>
      <div className="price-container">
        <div className="price-card">
          <div className="prices-container">
            <h4>
              <b>Starter Plan</b>
            </h4>
            <h3>
              <span style={{ fontSize: "70px", fontWeight: "bold" }}>Free</span>
            </h3>
            <p>No credit card required</p>
            <hr style={{ border: "1px solid grey", margin: "20px 0" }} />
            <ul>
              <li>
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: "black", marginRight: "8px" }}
                />
                Unlimited chats & websites
              </li>
              <li>
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: "black", marginRight: "8px" }}
                />
                60-day conversations history
              </li>
              <li>
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: "black", marginRight: "8px" }}
                />
                Basic widget customization
              </li>
            </ul>
            <div className="price-button">
              <VbitezButton>Purchase Now</VbitezButton>
              <span>Learn More</span>
            </div>
          </div>
        </div>
        <div className="price-card popular">
          <div className="prices-container">
            <h4>
              <b>Most Popular</b>
            </h4>
            <h3>
              ${" "}
              <span style={{ fontSize: "70px", fontWeight: "bold" }}>199</span>
              /month
            </h3>
            <p>No credit card required</p>

            <hr style={{ border: "1px solid grey", margin: "20px 0" }} />
            <ul>
              <li>
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: "black", marginRight: "8px" }}
                />
                Unlimited chats & websites
              </li>
              <li>
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: "black", marginRight: "8px" }}
                />
                60-day conversations history
              </li>
              <li>
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: "black", marginRight: "8px" }}
                />
                Basic widget customization
              </li>
            </ul>
            <div className="price-button">
              <VbitezButton>Purchase Now</VbitezButton>
              <span>Learn More</span>
            </div>
          </div>
        </div>
        <div className="price-card">
          <div className="prices-container">
            <h4>
              <b>Standard Plan</b>
            </h4>
            <h3>
              ${" "}
              <span style={{ fontSize: "70px", fontWeight: "bold" }}>400</span>
              /month
            </h3>
            <p>No credit card required</p>

            <hr style={{ border: "1px solid grey", margin: "20px 0" }} />
            <ul>
              <li>
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: "black", marginRight: "8px" }}
                />
                Unlimited chats & websites
              </li>
              <li>
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: "black", marginRight: "8px" }}
                />
                60-day conversations history
              </li>
              <li>
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: "black", marginRight: "8px" }}
                />
                Basic widget customization
              </li>
            </ul>
            <div className="price-button">
              <VbitezButton>Purchase Now</VbitezButton>
              <span>Learn More</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Price;
/*
  File Created By: Abed El-Fattah Amouneh
  Last Edited: Sunday 12 May 2024
  Last Edited by: Abed El-Fattah Amouneh
*/
