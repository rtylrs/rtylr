#include "ops/sepia.hpp"
#include "ops/band_util.hpp"

namespace voxire {
namespace ops {

vips::VImage sepia(vips::VImage img) {
    Split s = splitAlpha(img);
    // Desaturate to luminance, then multiply by a warm tone. Components above
    // 255 saturate on the uchar cast at encode time, which is the desired look.
    vips::VImage gray = s.color.colourspace(VIPS_INTERPRETATION_B_W)
                               .colourspace(VIPS_INTERPRETATION_sRGB);
    vips::VImage out = gray * std::vector<double>{1.07, 0.74, 0.43};
    return joinAlpha(out, s);
}

} // namespace ops
} // namespace voxire
