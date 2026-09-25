#include "ops/modulate.hpp"
#include "ops/band_util.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage modulate(vips::VImage img, double brightnessFactor,
                      double saturationFactor, double hueDegrees) {
    if (brightnessFactor < 0 || saturationFactor < 0) {
        throw ProcessingError("modulate brightness/saturation must not be negative");
    }
    Split s = splitAlpha(img);
    // In LCh: L = lightness, C = chroma, h = hue angle. Scale L and C, rotate h.
    vips::VImage lch = s.color.colourspace(VIPS_INTERPRETATION_LCH);
    lch = lch * std::vector<double>{brightnessFactor, saturationFactor, 1.0};
    lch = lch + std::vector<double>{0.0, 0.0, hueDegrees};
    vips::VImage out = lch.colourspace(VIPS_INTERPRETATION_sRGB);
    return joinAlpha(out, s);
}

} // namespace ops
} // namespace voxire
