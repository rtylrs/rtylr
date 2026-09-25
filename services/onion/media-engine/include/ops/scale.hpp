#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// scaleBy resizes the image by a uniform factor (e.g. 0.5 = half, 2.0 =
// double), preserving aspect ratio. This backs Cloudinary's percentage scale
// and DPR multipliers. factor must be > 0.
vips::VImage scaleBy(vips::VImage img, double factor);

} // namespace ops
} // namespace voxire
