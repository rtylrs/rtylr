#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// threshold converts the image to a 1-bit-style black/white result: pixels with
// luminance >= level (0-255) become white, the rest black. The output is a
// 3-band sRGB image. level must be in 0-255.
vips::VImage threshold(vips::VImage img, int level);

} // namespace ops
} // namespace voxire
