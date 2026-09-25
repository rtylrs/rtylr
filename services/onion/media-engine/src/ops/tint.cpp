#include "ops/tint.hpp"
#include "ops/band_util.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage tint(vips::VImage img, const Color& color, double amount) {
    if (amount < 0.0 || amount > 1.0) {
        throw ProcessingError("tint amount must be between 0 and 1");
    }
    Split s = splitAlpha(img);

    // Luminance in 0..1, replicated to three bands, then scaled onto the
    // target color to form the duotone.
    vips::VImage lum = s.color.colourspace(VIPS_INTERPRETATION_B_W)
                              .extract_band(0)
                              .linear(1.0 / 255.0, 0.0);
    vips::VImage lum3 = lum.bandjoin(lum).bandjoin(lum);
    vips::VImage duotone = lum3 * std::vector<double>{
        static_cast<double>(color.r),
        static_cast<double>(color.g),
        static_cast<double>(color.b),
    };

    // Blend: original*(1-amount) + duotone*amount.
    vips::VImage out =
        s.color.linear(1.0 - amount, 0.0) + duotone.linear(amount, 0.0);
    return joinAlpha(out, s);
}

} // namespace ops
} // namespace voxire
