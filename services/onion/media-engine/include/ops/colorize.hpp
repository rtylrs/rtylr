#pragma once

#include "../color.hpp"
#include <vips/vips8>

namespace voxire {
namespace ops {

// colorize blends the image toward a solid color by `amount` (0 = unchanged,
// 1 = the flat color), like Cloudinary e_colorize. amount must be in 0..1.
// Alpha is preserved.
vips::VImage colorize(vips::VImage img, const Color& color, double amount);

} // namespace ops
} // namespace voxire
