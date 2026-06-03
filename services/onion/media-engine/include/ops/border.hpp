#pragma once

#include "../color.hpp"
#include <vips/vips8>

namespace voxire {
namespace ops {

// border draws a solid border of the given width (px) and color around the
// image, growing the canvas by 2*width in each dimension. width must be > 0.
vips::VImage border(vips::VImage img, int width, const Color& color);

} // namespace ops
} // namespace voxire
