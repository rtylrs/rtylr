#include "analyze.hpp"

namespace voxire {

static std::string interpretationName(VipsInterpretation interp) {
    switch (interp) {
        case VIPS_INTERPRETATION_sRGB:   return "srgb";
        case VIPS_INTERPRETATION_RGB:    return "rgb";
        case VIPS_INTERPRETATION_B_W:    return "b-w";
        case VIPS_INTERPRETATION_GREY16: return "grey16";
        case VIPS_INTERPRETATION_CMYK:   return "cmyk";
        case VIPS_INTERPRETATION_LAB:    return "lab";
        case VIPS_INTERPRETATION_LCH:    return "lch";
        default:                         return "unknown";
    }
}

ImageInfo probeInfo(vips::VImage img) {
    ImageInfo info;
    info.width  = img.width();
    info.height = img.height();
    info.bands  = img.bands();
    info.hasAlpha = img.has_alpha();
    info.interpretation = interpretationName(img.interpretation());
    return info;
}

std::string infoToJSON(const ImageInfo& info) {
    return std::string("{") +
        "\"width\":" + std::to_string(info.width) + "," +
        "\"height\":" + std::to_string(info.height) + "," +
        "\"bands\":" + std::to_string(info.bands) + "," +
        "\"hasAlpha\":" + (info.hasAlpha ? "true" : "false") + "," +
        "\"interpretation\":\"" + info.interpretation + "\"" +
        "}";
}

} // namespace voxire
