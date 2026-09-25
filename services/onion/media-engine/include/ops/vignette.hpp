#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// vignette darkens the image toward the edges using a radial gradient.
// strength is 0..1 (0 = no effect, 1 = edges fully black). Alpha is preserved.
vips::VImage vignette(vips::VImage img, double strength);

} // namespace ops
} // namespace voxire
