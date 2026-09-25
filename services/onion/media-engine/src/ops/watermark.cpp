#include "ops/watermark.hpp"
#include "ops/composite.hpp"
#include "ops/opacity.hpp"

namespace voxire {
namespace ops {

vips::VImage watermark(vips::VImage base, vips::VImage mark,
                       Gravity gravity, int margin, double opacity) {
    // Reuse the opacity op to fade the mark, then the composite op to place it.
    if (opacity < 1.0) {
        mark = ops::opacity(mark, opacity);
    }
    return composite(base, mark, gravity, margin);
}

} // namespace ops
} // namespace voxire
