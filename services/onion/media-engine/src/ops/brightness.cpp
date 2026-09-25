#include "ops/brightness.hpp"
#include "ops/band_util.hpp"

namespace voxire {
namespace ops {

vips::VImage brightness(vips::VImage img, double delta) {
    // out = color * 1 + delta, applied to color bands only so alpha is intact.
    Split s = splitAlpha(img);
    vips::VImage out = s.color.linear(1.0, delta);
    return joinAlpha(out, s);
}

} // namespace ops
} // namespace voxire
