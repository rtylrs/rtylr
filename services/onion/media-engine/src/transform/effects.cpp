#include "transform/effects.hpp"
#include "common.hpp"
#include "numeric.hpp"

#include "ops/blur.hpp"
#include "ops/sharpen.hpp"
#include "ops/grayscale.hpp"
#include "ops/sepia.hpp"
#include "ops/invert.hpp"
#include "ops/pixelate.hpp"
#include "ops/brightness.hpp"
#include "ops/contrast.hpp"
#include "ops/saturation.hpp"
#include "ops/posterize.hpp"
#include "ops/threshold.hpp"
#include "ops/median.hpp"
#include "ops/vignette.hpp"

namespace voxire {
namespace transform {

// numOr returns the parsed value, or fallback when value is empty.
static double numOr(const std::string& value, double fallback) {
    if (value.empty()) return fallback;
    return parseDoubleStrict(value, "effect value");
}

vips::VImage applyEffect(vips::VImage img, const std::string& name,
                         const std::string& value) {
    if (name == "blur") {
        // Cloudinary strength ~1..2000; map to a libvips sigma.
        return ops::blur(img, numOr(value, 300.0) / 100.0);
    }
    if (name == "sharpen") {
        return ops::sharpen(img, numOr(value, 100.0) / 100.0);
    }
    if (name == "grayscale" || name == "greyscale") {
        return ops::grayscale(img);
    }
    if (name == "sepia") {
        return ops::sepia(img);
    }
    if (name == "negate") {
        return ops::invert(img);
    }
    if (name == "pixelate") {
        return ops::pixelate(img, static_cast<int>(numOr(value, 8.0)));
    }
    if (name == "brightness") {
        // -99..100 percent -> 0..255 delta.
        return ops::brightness(img, numOr(value, 0.0) * 2.55);
    }
    if (name == "contrast") {
        return ops::contrast(img, 1.0 + numOr(value, 0.0) / 100.0);
    }
    if (name == "saturation") {
        return ops::saturation(img, 1.0 + numOr(value, 0.0) / 100.0);
    }
    if (name == "posterize") {
        return ops::posterize(img, clampInt(static_cast<int>(numOr(value, 4.0)), 2, 255));
    }
    if (name == "blackwhite") {
        return ops::threshold(img, clampInt(static_cast<int>(numOr(value, 128.0)), 0, 255));
    }
    if (name == "oil_paint") {
        int size = static_cast<int>(numOr(value, 5.0));
        if (size % 2 == 0) size += 1; // median needs an odd window
        return ops::median(img, size < 3 ? 3 : size);
    }
    if (name == "vignette") {
        // value is 0..100 strength.
        return ops::vignette(img, clampDouble(numOr(value, 30.0) / 100.0, 0.0, 1.0));
    }
    throw ProcessingError("unknown effect: " + name);
}

} // namespace transform
} // namespace voxire
