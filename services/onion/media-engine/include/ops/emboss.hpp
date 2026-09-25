#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// emboss applies a 3x3 emboss convolution, giving the image a raised, engraved
// look. Alpha is preserved.
vips::VImage emboss(vips::VImage img);

} // namespace ops
} // namespace voxire
