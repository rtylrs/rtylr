#include "ops/emboss.hpp"
#include "ops/band_util.hpp"

namespace voxire {
namespace ops {

vips::VImage emboss(vips::VImage img) {
    Split s = splitAlpha(img);
    // 3x3 emboss kernel with a +1 center so flat areas stay mid-toned.
    vips::VImage kernel = vips::VImage::new_matrixv(3, 3,
        -2.0, -1.0, 0.0,
        -1.0,  1.0, 1.0,
         0.0,  1.0, 2.0);
    vips::VImage out = s.color.conv(kernel);
    return joinAlpha(out, s);
}

} // namespace ops
} // namespace voxire
