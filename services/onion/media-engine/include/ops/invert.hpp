#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// invert produces the photographic negative of the image, leaving any alpha
// channel untouched.
vips::VImage invert(vips::VImage img);

} // namespace ops
} // namespace voxire
