#include "ops/colorize.hpp"
#include "ops/band_util.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage colorize(vips::VImage img, const Color& color, double amount) {
    if (amount < 0.0 || amount > 1.0) {
        throw ProcessingError("colorize amount must be between 0 and 1");
    }
    Split s = splitAlpha(img);
    // A flat field of the target color, blended with the original:
    //   out = original*(1-amount) + color*amount.
    vips::VImage solid = s.color.new_from_image(std::vector<double>{
        static_cast<double>(color.r),
        static_cast<double>(color.g),
        static_cast<double>(color.b),
    });
    vips::VImage out =
        s.color.linear(1.0 - amount, 0.0) + solid.linear(amount, 0.0);
    return joinAlpha(out, s);
}

} // namespace ops
} // namespace voxire
