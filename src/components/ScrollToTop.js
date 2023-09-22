import { useEffect } from "react";
import { useLocation } from "react-router-dom";

export default function ScrollToTop() {
    const { pathname } = useLocation(); // Get the current path using the `useLocation` hook

    useEffect(() => {
        const scrollToTop = () => {
            window.scrollTo({ top: 0, behavior: "smooth" }); // Scroll to the top of the page with smooth behavior
        };

        scrollToTop(); // Scroll to the top on path change

        // Cleanup function to remove the event listener when the component unmounts or when the path changes
        return () => {
            window.removeEventListener("scroll", scrollToTop);
        };
    }, [pathname]);

    return null; // Return null since this component doesn't render anything
}