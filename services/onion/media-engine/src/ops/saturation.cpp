#include "ops/saturation.hpp"
#include "ops/band_util.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage saturation(vips::VImage img, double factor) {
    if (factor < 0) {
        throw ProcessingError("saturation factor must not be negative");
    }
    Split s = splitAlpha(img);
    // In LCh the middle band is chroma; scaling it changes saturation while
    // leaving lightness (L) and hue (h) intact.
    vips::VImage lch = s.color.colourspace(VIPS_INTERPRETATION_LCH);
    lch = lch * std::vector<double>{1.0, factor, 1.0};
    vips::VImage out = lch.colourspace(VIPS_INTERPRETATION_sRGB);
    return joinAlpha(out, s);
}

} // namespace ops
} // namespace voxire
