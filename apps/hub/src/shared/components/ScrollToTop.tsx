import { useEffect } from "react";
import { useLocation } from "react-router-dom";

const ScrollToTop: React.FC = () => {
  const { pathname } = useLocation(); // Get the current path using the `useLocation` hook

  useEffect(() => {
    const scrollToTop = () => {
      window.scrollTo({ top: 0, behavior: "smooth" }); // Scroll to the top of the page with smooth behavior
    };

    scrollToTop(); // Scroll to the top on path change

    return () => {
      window.removeEventListener("scroll", scrollToTop); // Remove the event listener
    };
  }, [pathname]);

  return null; // Return null since this component doesn't render anything
};

export default ScrollToTop;
