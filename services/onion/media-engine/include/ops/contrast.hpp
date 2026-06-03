#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// contrast scales each color channel about the 128 midpoint by factor. factor
// > 1 increases contrast, between 0 and 1 reduces it. factor must be > 0.
// Alpha is preserved.
vips::VImage contrast(vips::VImage img, double factor);

} // namespace ops
} // namespace voxire
