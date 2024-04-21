import { useTheme } from "../HOCs/ThemeContext";

export default function useColors() {
  const { isDarkMode } = useTheme();

  const darkThemeColors = {
    primary: "#05A79A",
    secondary: "#13CDBE",
    darkColor: "#016B63",
    background: "#0F0F0F",
    bgColor: "#0F0F0F",
    light: "#fff",
    dark: "#000",
    itemBackground: "#373737",
    textColor: "#fff",
    subTextColor: "#B0B0B0",
    lightGray: "#D9D9D9",
  };

  const lightThemeColors = {
    primary: "#05A79A",
    secondary: "#13CDBE",
    darkColor: "#016B63",
    background: "#fff",
    bgColor: "#EEEEEE",
    light: "#fff",
    dark: "#000",
    itemBackground: "#f8f8f8",
    textColor: "#373737",
    subTextColor: "#B0B0B0",
    lightGray: "#D9D9D9",
  };

  switch (isDarkMode ? "dark" : "light") {
    case "dark":
      return darkThemeColors;
    case "light":
      return lightThemeColors;
    default:
      return lightThemeColors;
  }
}
