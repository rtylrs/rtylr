#pragma once

#include "../color.hpp"
#include <vips/vips8>

namespace voxire {
namespace ops {

// averageColor returns the mean color of the image (the Cloudinary
// predominant/average use case, handy for placeholder backgrounds). Alpha is
// ignored.
Color averageColor(vips::VImage img);

} // namespace ops
} // namespace voxire
