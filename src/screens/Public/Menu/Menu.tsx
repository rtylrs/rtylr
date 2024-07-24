import {
  faComments,
  faPhone,
  faUtensils,
  faHome,
  faInfoCircle,
} from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import useAsync from "../../../hooks/useAsync";
import menuApi from "../../../api/menu";
import LandingSection from "./Sections/LandingSection";
import MenuSection from "./Sections/MenuSection";
import AboutSection from "./Sections/AboutSection";
import ContactSection from "./Sections/ContactSection";
import FeedbackSection from "./Sections/FeedbackSection";

function Menu() {
  const [menuArray, setmenuArray] = useState([]);
  const [restaurant, setRestaurant] = useState<any>();
  const [theme, setTheme] = useState<any>();
  const [status, setStatus] = useState("idle");
  const [retryCount, setRetryCount] = useState<number>(0);
  const [activeSection, setActiveSection] = useState<string>("menu");
  const { id } = useParams();

  const getMenu = useAsync({
    fn: menuApi.getMenu,
    onSuccess: () => {
      setRetryCount(0);
      const { details } = getMenu.result;
      setmenuArray(details.menu.items);
      setRestaurant(details.business);
      setTheme(details.theme);
      setStatus("succeeded");

      document.title = details?.business?.name || "Menu";

      const url = new URL(window.location.href);

      if (!url.searchParams.has("title")) {
        url.searchParams.append("title", details?.business?.name || "");
      }

      window.history.replaceState({}, "", url.toString());
    },
    onError: () => {
      setStatus("failed");

      if (retryCount <= 3) {
        setTimeout(() => {
          getMenu.main({ id });
          setRetryCount(retryCount + 1);
        }, 10000);
      }
    },
    callOnMount: false,
  });

  useEffect(() => {
    setStatus("loading");
    getMenu.main({ id });

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (status === "loading") {
    return <div className="menu-container">Loading...</div>;
  }

  if (status === "failed") {
    return <div className="menu-container">Error loading menu.</div>;
  }

  const renderActiveSection = () => {
    switch (activeSection) {
      case "landing":
        return <LandingSection />;
      case "menu":
        return <MenuSection menuArray={menuArray} />;
      case "about":
        return <AboutSection />;
      case "contact":
        return <ContactSection />;
      case "feedback":
        return <FeedbackSection />;
      default:
        return null;
    }
  };

  return (
    <div
      className="menu-container"
      style={{ backgroundColor: theme?.menuBackgroundColor }}
    >
      <div className="menu">
        <h1 style={{ color: theme?.primaryColor }}>
          {restaurant?.name}'s Menu
        </h1>
        {renderActiveSection()}
      </div>
      <div className="bottom-navigation">
        <div onClick={() => setActiveSection("about")}>
          <FontAwesomeIcon icon={faInfoCircle} />
        </div>
        {restaurant?.addons?.landing && (
          <div onClick={() => setActiveSection("menu")}>
            <FontAwesomeIcon icon={faUtensils} />
          </div>
        )}
        <div
          onClick={() => {
            if (restaurant?.addons?.landing) {
              setActiveSection("landing");
            } else {
              setActiveSection("menu");
            }
          }}
        >
          <FontAwesomeIcon
            icon={restaurant?.addons.landing ? faHome : faUtensils}
          />
        </div>
        <div onClick={() => setActiveSection("contact")}>
          <FontAwesomeIcon icon={faPhone} />
        </div>
        <div onClick={() => setActiveSection("feedback")}>
          <FontAwesomeIcon icon={faComments} />
        </div>
      </div>
    </div>
  );
}

export default Menu;
