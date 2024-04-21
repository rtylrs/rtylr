import { Dimensions, PixelRatio } from "react-native";

export const windowWidth = Dimensions.get("window").width;
export const windowHeight = Dimensions.get("window").height;

export const scaleFont = (size) => {
  const scaleWidth = windowWidth / 375;
  const scaleHeight = windowHeight / 667;
  const scale = Math.min(scaleWidth, scaleHeight);
  return Math.round(PixelRatio.roundToNearestPixel(size * scale));
};
