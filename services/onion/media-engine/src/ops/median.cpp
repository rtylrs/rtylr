#include "ops/median.hpp"
#include "ops/band_util.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage median(vips::VImage img, int size) {
    if (size < 3 || size % 2 == 0) {
        throw ProcessingError("median window size must be an odd number >= 3");
    }
    Split s = splitAlpha(img);
    // A rank filter that picks the middle element of each window is the median.
    const int index = (size * size) / 2;
    vips::VImage out = s.color.rank(size, size, index);
    return joinAlpha(out, s);
}

} // namespace ops
} // namespace voxire
