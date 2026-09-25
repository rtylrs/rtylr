#pragma once

#include "../color.hpp"
#include <vips/vips8>

namespace voxire {
namespace ops {

// flatten composites a transparent image over a solid background, removing the
// alpha channel (the Cloudinary b_<color> background). A fully opaque image is
// returned unchanged.
vips::VImage flatten(vips::VImage img, const Color& background);

} // namespace ops
} // namespace voxire
