#include "ops/grayscale.hpp"

namespace voxire {
namespace ops {

vips::VImage grayscale(vips::VImage img) {
    const bool alpha = img.has_alpha();
    vips::VImage extra;
    if (alpha) {
        const int n = img.bands();
        extra = img.extract_band(n - 1);
        img = img.extract_band(0, vips::VImage::option()->set("n", n - 1));
    }
    // Desaturate via the perceptual B_W space, then return to sRGB.
    vips::VImage out = img.colourspace(VIPS_INTERPRETATION_B_W)
                          .colourspace(VIPS_INTERPRETATION_sRGB);
    if (alpha) {
        out = out.bandjoin(extra);
    }
    return out;
}

} // namespace ops
} // namespace voxire
