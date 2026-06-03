#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// saturation scales color intensity by factor in the LCh colorspace. factor > 1
// is more vivid, 0 is fully desaturated (grayscale). factor must be >= 0.
// Alpha is preserved.
vips::VImage saturation(vips::VImage img, double factor);

} // namespace ops
} // namespace voxire
