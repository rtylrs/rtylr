#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// edgeDetect highlights edges with a 3x3 Laplacian convolution, producing a
// line-drawing effect. Alpha is preserved.
vips::VImage edgeDetect(vips::VImage img);

} // namespace ops
} // namespace voxire
