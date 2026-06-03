#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// brightness shifts every color channel by delta (in 0-255 units; negative
// darkens, positive brightens). Alpha is preserved.
vips::VImage brightness(vips::VImage img, double delta);

} // namespace ops
} // namespace voxire
