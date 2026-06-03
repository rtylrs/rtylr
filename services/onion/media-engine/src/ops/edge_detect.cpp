#include "ops/edge_detect.hpp"
#include "ops/band_util.hpp"

namespace voxire {
namespace ops {

vips::VImage edgeDetect(vips::VImage img) {
    Split s = splitAlpha(img);
    // Laplacian kernel — responds to intensity changes in all directions.
    vips::VImage kernel = vips::VImage::new_matrixv(3, 3,
        0.0, -1.0, 0.0,
       -1.0,  4.0, -1.0,
        0.0, -1.0, 0.0);
    vips::VImage out = s.color.conv(kernel);
    return joinAlpha(out, s);
}

} // namespace ops
} // namespace voxire
