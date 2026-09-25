#include "ops/posterize.hpp"
#include "ops/band_util.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage posterize(vips::VImage img, int levels) {
    if (levels < 2 || levels > 255) {
        throw ProcessingError("posterize levels must be between 2 and 255");
    }
    const double step = 256.0 / levels;
    Split s = splitAlpha(img);
    // Quantise: floor(v / step) * step + step/2 lands each band in the middle
    // of its bucket.
    vips::VImage out = s.color.linear(1.0 / step, 0.0)
                              .floor()
                              .linear(step, step / 2.0);
    return joinAlpha(out, s);
}

} // namespace ops
} // namespace voxire
