#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// median applies a square median (rank) filter of the given odd window size,
// removing speckle/salt-and-pepper noise while keeping edges. size must be an
// odd number >= 3. Alpha is preserved.
vips::VImage median(vips::VImage img, int size);

} // namespace ops
} // namespace voxire
