#include "ops/pixelate.hpp"
#include "common.hpp"
#include <algorithm>

namespace voxire {
namespace ops {

vips::VImage pixelate(vips::VImage img, int blockSize) {
    if (blockSize < 2) {
        throw ProcessingError("pixelate block size must be at least 2");
    }
    const int w = img.width();
    const int h = img.height();

    // Shrink by the block factor (averaging each block to one pixel), then
    // zoom back with nearest-neighbour so each block is a flat color.
    vips::VImage small = img.shrink(blockSize, blockSize);
    vips::VImage big = small.zoom(blockSize, blockSize);

    // zoom rounds up; trim back to the original dimensions.
    const int cw = std::min(w, big.width());
    const int ch = std::min(h, big.height());
    return big.crop(0, 0, cw, ch);
}

} // namespace ops
} // namespace voxire
