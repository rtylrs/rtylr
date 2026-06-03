#pragma once

#include "../geometry.hpp"
#include <vips/vips8>

namespace voxire {
namespace ops {

// crop extracts the box r from img. The box is clamped to the image bounds; an
// out-of-range origin throws ProcessingError.
vips::VImage crop(vips::VImage img, const Box& r);

} // namespace ops
} // namespace voxire
