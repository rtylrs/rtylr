#include "ops/flip.hpp"

namespace voxire {
namespace ops {

vips::VImage flipVertical(vips::VImage img) {
    return img.flip(VIPS_DIRECTION_VERTICAL);
}

} // namespace ops
} // namespace voxire
