import { Button } from '@voxire/components';
import { FaStar } from "react-icons/fa";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import React from "react";
import { useSelector } from "react-redux";
import { selectIsAuthenticated } from "../../../../redux/reducers/userSlice";

const Price: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const isAuthenticated = useSelector(selectIsAuthenticated);

  return (
    <div className="Prices">
      <h1 className="rtylr-title price-text">{t("PricingSection.title")}</h1>
      <p className="rtylr-text-price-text">
        {t("PricingSection.description")}
      </p>
      <div className="price-container">
        <div className="price-card">
          <div className="prices-container">
            <h4>
              <b>{t("PricingSection.Section1_title")}</b>
            </h4>
            <h3>
              <span style={{ fontSize: "70px", fontWeight: "bold" }}>
                {t("PricingSection.Section1_Price")}
              </span>
            </h3>
            <p>{t("PricingSection.Section1_Requirements")}</p>
            <hr style={{ border: "1px solid grey", margin: "20px 0" }} />
            <ul>
              <li>
                <FaStar
                  style={{ color: "black", marginRight: "8px" }}
                />
                {t("PricingSection.Section1_star1")}
              </li>
              <li>
                <FaStar
                  style={{ color: "black", marginRight: "8px" }}
                />
                {t("PricingSection.Section1_star2")}
              </li>
              <li>
                <FaStar
                  style={{ color: "black", marginRight: "8px" }}
                />
                Basic widget customization
              </li>
            </ul>
            <div className="price-button">
              <span
                onClick={() => {
                  navigate("/pricing");
                }}
              >
                Learn More
              </span>
            </div>
          </div>
        </div>
        <div className="price-card popular">
          <div className="prices-container">
            <h4>
              <b>Advanced</b>
            </h4>
            <h3>
              $ <span style={{ fontSize: "70px", fontWeight: "bold" }}>49</span>
              /year
            </h3>
            <p>For small restaurant or bar.</p>

            <hr style={{ border: "1px solid grey", margin: "20px 0" }} />
            <ul>
              <li>
                <FaStar
                  style={{ color: "black", marginRight: "8px" }}
                />
                Unlimited chats & websites
              </li>
              <li>
                <FaStar
                  style={{ color: "black", marginRight: "8px" }}
                />
                60-day conversations history
              </li>
              <li>
                <FaStar
                  style={{ color: "black", marginRight: "8px" }}
                />
                Basic widget customization
              </li>
            </ul>
            <div className="price-button">
              {/* <Button>Get Started</Button> */}
              <span
                onClick={() => {
                  navigate("/pricing");
                }}
              >
                Learn More
              </span>
            </div>
          </div>
        </div>
        <div className="price-card">
          <div className="prices-container">
            <h4>
              <b>Premium</b>
            </h4>
            <h3>
              $ <span style={{ fontSize: "70px", fontWeight: "bold" }}>59</span>
              /year
            </h3>
            <p>For high-demand restaurants and bars.</p>

            <hr style={{ border: "1px solid grey", margin: "20px 0" }} />
            <ul>
              <li>
                <FaStar
                  style={{ color: "black", marginRight: "8px" }}
                />
                Unlimited chats & websites
              </li>
              <li>
                <FaStar
                  style={{ color: "black", marginRight: "8px" }}
                />
                60-day conversations history
              </li>
              <li>
                <FaStar
                  style={{ color: "black", marginRight: "8px" }}
                />
                Basic widget customization
              </li>
            </ul>
            <div className="price-button">
              {/* <Button>Get Started</Button> */}
              <span
                onClick={() => {
                  navigate("/pricing");
                }}
              >
                Learn More
              </span>
            </div>
          </div>
        </div>
      </div>
      <div
        style={{
          marginTop: "50px",
          marginBottom: "0px",
        }}
      >
        <Button
          onClick={() => {
            if (isAuthenticated) {
              navigate("/app/business");
            } else {
              navigate("/auth/register");
            }
          }}
        >
          {isAuthenticated ? "Dashboard" : "Get Started"}
        </Button>
      </div>
    </div>
  );
};

export default Price;
