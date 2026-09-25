#pragma once

#include "../gravity.hpp"
#include <vips/vips8>

namespace voxire {
namespace ops {

// composite places `overlay` over `base` using normal (over) alpha blending,
// anchored by `gravity` and inset from the edges by `margin` pixels. The
// overlay is positioned, not scaled; size it before compositing.
vips::VImage composite(vips::VImage base, vips::VImage overlay,
                       Gravity gravity, int margin);

} // namespace ops
} // namespace voxire
