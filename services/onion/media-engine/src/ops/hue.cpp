#include "ops/hue.hpp"
#include "ops/band_util.hpp"

namespace voxire {
namespace ops {

vips::VImage hue(vips::VImage img, double degrees) {
    Split s = splitAlpha(img);
    // LCh's third band is hue angle; adding to it rotates the color wheel.
    vips::VImage lch = s.color.colourspace(VIPS_INTERPRETATION_LCH);
    lch = lch + std::vector<double>{0.0, 0.0, degrees};
    vips::VImage out = lch.colourspace(VIPS_INTERPRETATION_sRGB);
    return joinAlpha(out, s);
}

} // namespace ops
} // namespace voxire
