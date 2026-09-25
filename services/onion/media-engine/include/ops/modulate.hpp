#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// modulate adjusts brightness, saturation, and hue together in one LCh pass
// (like ImageMagick -modulate / Cloudinary combined effects). brightnessFactor
// and saturationFactor scale lightness and chroma (1 = unchanged); hueDegrees
// rotates the hue. Alpha is preserved.
vips::VImage modulate(vips::VImage img, double brightnessFactor,
                      double saturationFactor, double hueDegrees);

} // namespace ops
} // namespace voxire
