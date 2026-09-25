#include "ops/average_color.hpp"
#include "ops/band_util.hpp"
#include "numeric.hpp"

namespace voxire {
namespace ops {

Color averageColor(vips::VImage img) {
    // Drop alpha, force sRGB, then shrink to a single pixel — its value is the
    // mean color across the whole image.
    Split s = splitAlpha(img);
    vips::VImage rgb = s.color.colourspace(VIPS_INTERPRETATION_sRGB);
    vips::VImage one = rgb.resize(1.0 / rgb.width(),
        vips::VImage::option()->set("vscale", 1.0 / rgb.height()));

    std::vector<double> px = one.getpoint(0, 0);
    Color c;
    if (px.size() > 0) c.r = clampInt(static_cast<int>(px[0] + 0.5), 0, 255);
    if (px.size() > 1) c.g = clampInt(static_cast<int>(px[1] + 0.5), 0, 255);
    if (px.size() > 2) c.b = clampInt(static_cast<int>(px[2] + 0.5), 0, 255);
    return c;
}

} // namespace ops
} // namespace voxire
