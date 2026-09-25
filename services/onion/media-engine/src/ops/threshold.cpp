#include "ops/threshold.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage threshold(vips::VImage img, int level) {
    if (level < 0 || level > 255) {
        throw ProcessingError("threshold level must be between 0 and 255");
    }
    // Reduce to a single luminance band, compare against the level (vips
    // relational ops yield 255 for true / 0 for false), then expand to sRGB.
    vips::VImage mono = img.colourspace(VIPS_INTERPRETATION_B_W).extract_band(0);
    vips::VImage mask = mono.relational_const(
        VIPS_OPERATION_RELATIONAL_MOREEQ,
        std::vector<double>{static_cast<double>(level)});
    return mask.colourspace(VIPS_INTERPRETATION_sRGB);
}

} // namespace ops
} // namespace voxire
