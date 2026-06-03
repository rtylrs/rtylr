#pragma once

#include "../gravity.hpp"
#include <vips/vips8>

namespace voxire {
namespace ops {

// watermark overlays `mark` onto `base` at the given gravity and margin, after
// scaling the mark's alpha by `opacity` (0..1) so it can be made translucent.
// The mark is positioned, not resized.
vips::VImage watermark(vips::VImage base, vips::VImage mark,
                       Gravity gravity, int margin, double opacity);

} // namespace ops
} // namespace voxire
