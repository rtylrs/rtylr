#include "ops/rounded_corners.hpp"
#include "ops/band_util.hpp"
#include "common.hpp"
#include <string>

namespace voxire {
namespace ops {

vips::VImage roundedCorners(vips::VImage img, int radius) {
    if (radius <= 0) {
        throw ProcessingError("corner radius must be greater than 0");
    }
    const int w = img.width();
    const int h = img.height();

    // Render a white rounded rectangle as the mask via the SVG loader.
    const std::string svg =
        "<svg xmlns='http://www.w3.org/2000/svg' width='" + std::to_string(w) +
        "' height='" + std::to_string(h) + "'>"
        "<rect x='0' y='0' width='" + std::to_string(w) +
        "' height='" + std::to_string(h) +
        "' rx='" + std::to_string(radius) + "' ry='" + std::to_string(radius) +
        "' fill='#ffffff'/></svg>";

    vips::VImage mask =
        vips::VImage::new_from_buffer(svg.c_str(), svg.size(), "");
    vips::VImage maskAlpha = mask.extract_band(0);

    // Replace any existing alpha with the rounded mask over the color bands.
    Split s = splitAlpha(img);
    return s.color.bandjoin(maskAlpha);
}

} // namespace ops
} // namespace voxire
