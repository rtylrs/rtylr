#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// sepia desaturates the image and applies a warm brown tone, like the
// Cloudinary e_sepia effect. Alpha is preserved.
vips::VImage sepia(vips::VImage img);

} // namespace ops
} // namespace voxire
