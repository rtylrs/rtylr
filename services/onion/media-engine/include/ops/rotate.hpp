#pragma once

#include "../color.hpp"
#include <vips/vips8>

namespace voxire {
namespace ops {

// rotate turns the image clockwise by an arbitrary angle in degrees. Multiples
// of 90 use the lossless rot operator; other angles use a similarity transform
// and fill the exposed corners with the given background color.
vips::VImage rotate(vips::VImage img, double degrees, const Color& background);

} // namespace ops
} // namespace voxire
