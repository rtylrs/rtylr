#include "ops/rotate.hpp"
#include <cmath>

namespace voxire {
namespace ops {

vips::VImage rotate(vips::VImage img, double degrees, const Color& background) {
    // Normalise into [0, 360).
    double angle = std::fmod(degrees, 360.0);
    if (angle < 0) angle += 360.0;

    // Lossless right-angle rotations.
    if (angle == 0.0)   return img;
    if (angle == 90.0)  return img.rot(VIPS_ANGLE_D90);
    if (angle == 180.0) return img.rot(VIPS_ANGLE_D180);
    if (angle == 270.0) return img.rot(VIPS_ANGLE_D270);

    // Arbitrary angle: similarity transform with a colored background.
    return img.similarity(
        vips::VImage::option()
            ->set("angle", angle)
            ->set("background", background.bands())
    );
}

} // namespace ops
} // namespace voxire
