#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// posterize reduces each color channel to `levels` discrete steps, giving a
// banded, poster-like look (Cloudinary e_posterize). levels must be in 2..255.
// Alpha is preserved.
vips::VImage posterize(vips::VImage img, int levels);

} // namespace ops
} // namespace voxire
