#include "ops/normalize.hpp"
#include "ops/band_util.hpp"

namespace voxire {
namespace ops {

vips::VImage normalize(vips::VImage img) {
    Split s = splitAlpha(img);
    const double lo = s.color.min();
    const double hi = s.color.max();
    if (hi <= lo) {
        return img; // flat image — nothing to stretch
    }
    const double scale = 255.0 / (hi - lo);
    vips::VImage out = s.color.linear(scale, -lo * scale);
    return joinAlpha(out, s);
}

} // namespace ops
} // namespace voxire
