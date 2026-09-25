#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// hue rotates the color wheel by degrees (any value; wraps at 360) in the LCh
// colorspace, leaving lightness and saturation unchanged. Alpha is preserved.
vips::VImage hue(vips::VImage img, double degrees);

} // namespace ops
} // namespace voxire
