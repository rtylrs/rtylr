#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// opacity scales the image's alpha channel by factor (0 = fully transparent,
// 1 = unchanged), adding an opaque alpha first when the image has none. factor
// must be in 0..1.
vips::VImage opacity(vips::VImage img, double factor);

} // namespace ops
} // namespace voxire
