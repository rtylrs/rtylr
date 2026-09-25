#include "ops/flop.hpp"

namespace voxire {
namespace ops {

vips::VImage flopHorizontal(vips::VImage img) {
    return img.flip(VIPS_DIRECTION_HORIZONTAL);
}

} // namespace ops
} // namespace voxire
