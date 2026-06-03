#include "ops/contrast.hpp"
#include "ops/band_util.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage contrast(vips::VImage img, double factor) {
    if (factor <= 0) {
        throw ProcessingError("contrast factor must be greater than 0");
    }
    // out = (color - 128) * factor + 128 = color*factor + (128 - 128*factor),
    // applied to color bands only so alpha is untouched.
    const double offset = 128.0 - 128.0 * factor;
    Split s = splitAlpha(img);
    vips::VImage out = s.color.linear(factor, offset);
    return joinAlpha(out, s);
}

} // namespace ops
} // namespace voxire
