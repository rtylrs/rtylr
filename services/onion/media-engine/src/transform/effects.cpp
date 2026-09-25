#include "transform/effects.hpp"
#include "color.hpp"
#include "common.hpp"
#include "numeric.hpp"

#include <algorithm>

#include "ops/alpha.hpp"
#include "ops/blur.hpp"
#include "ops/brightness.hpp"
#include "ops/colorize.hpp"
#include "ops/contrast.hpp"
#include "ops/duotone.hpp"
#include "ops/edge_detect.hpp"
#include "ops/emboss.hpp"
#include "ops/flip.hpp"
#include "ops/flop.hpp"
#include "ops/gamma.hpp"
#include "ops/grayscale.hpp"
#include "ops/hue.hpp"
#include "ops/invert.hpp"
#include "ops/median.hpp"
#include "ops/modulate.hpp"
#include "ops/normalize.hpp"
#include "ops/pixelate.hpp"
#include "ops/saturation.hpp"
#include "ops/sepia.hpp"
#include "ops/sharpen.hpp"
#include "ops/posterize.hpp"
#include "ops/threshold.hpp"
#include "ops/tint.hpp"
#include "ops/trim.hpp"
#include "ops/vignette.hpp"

#include <sstream>
#include <string>
#include <vector>

namespace voxire {
namespace transform {

// numOr returns the parsed value, or fallback when value is empty.
static double numOr(const std::string& value, double fallback) {
    if (value.empty()) return fallback;
    return parseDoubleStrict(value, "effect value");
}

static std::vector<std::string> splitEffectArgs(const std::string& value) {
    std::vector<std::string> out;
    std::string cur;
    for (char c : value) {
        if (c == ':' || c == '|') {
            out.push_back(cur);
            cur.clear();
        } else {
            cur.push_back(c);
        }
    }
    out.push_back(cur);
    return out;
}

static int intOr(const std::string& value, int fallback) {
    if (value.empty()) return fallback;
    return parseIntStrict(value, "effect value");
}

static double percentOrUnit(const std::string& value, double fallback) {
    if (value.empty()) return fallback;
    double n = parseDoubleStrict(value, "effect percentage");
    if (n > 1.0) n = n / 100.0;
    return clampDouble(n, 0.0, 1.0);
}

vips::VImage applyEffect(vips::VImage img, const std::string& name,
                         const std::string& value) {
	if (name == "blur") {
		// Cloudinary strength ~1..2000; map to a libvips sigma.
		return ops::blur(img, std::clamp(numOr(value, 300.0), 1.0, 2000.0) / 100.0);
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
    if (name == "invert") {
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
    if (name == "hue") {
        return ops::hue(img, numOr(value, 0.0));
    }
    if (name == "gamma") {
        return ops::gamma(img, numOr(value, 1.0));
    }
    if (name == "modulate") {
        auto parts = splitEffectArgs(value);
        const double brightness = parts.size() > 0 && !parts[0].empty()
            ? parseDoubleStrict(parts[0], "modulate brightness") / 100.0 : 1.0;
        const double saturation = parts.size() > 1 && !parts[1].empty()
            ? parseDoubleStrict(parts[1], "modulate saturation") / 100.0 : 1.0;
        const double hue = parts.size() > 2 && !parts[2].empty()
            ? parseDoubleStrict(parts[2], "modulate hue") : 0.0;
        return ops::modulate(img, brightness, saturation, hue);
    }
    if (name == "normalize" || name == "auto_contrast" || name == "improve") {
        return ops::normalize(img);
    }
    if (name == "posterize") {
        return ops::posterize(img, clampInt(static_cast<int>(numOr(value, 4.0)), 2, 255));
    }
    if (name == "blackwhite") {
        return ops::threshold(img, clampInt(static_cast<int>(numOr(value, 128.0)), 0, 255));
    }
	if (name == "oil_paint") {
		int size = std::clamp(static_cast<int>(numOr(value, 5.0)), 3, 31);
		if (size % 2 == 0) size -= 1; // median needs an odd window
		return ops::median(img, size);
	}
	if (name == "median") {
		int size = std::clamp(intOr(value, 5), 3, 31);
		if (size % 2 == 0) size -= 1;
		return ops::median(img, size);
    }
    if (name == "trim") {
        return ops::trim(img, numOr(value, 10.0));
    }
    if (name == "edge_detect" || name == "outline") {
        return ops::edgeDetect(img);
    }
    if (name == "emboss") {
        return ops::emboss(img);
    }
    if (name == "vignette") {
        // value is 0..100 strength.
        return ops::vignette(img, clampDouble(numOr(value, 30.0) / 100.0, 0.0, 1.0));
    }
    if (name == "flip") {
        return ops::flipVertical(img);
    }
    if (name == "flop") {
        return ops::flopHorizontal(img);
    }
    if (name == "remove_alpha") {
        return ops::removeAlpha(img);
    }
    if (name == "alpha") {
        return ops::addAlpha(img, clampInt(intOr(value, 255), 0, 255));
    }
    if (name == "colorize") {
        auto parts = splitEffectArgs(value);
        if (parts.size() < 2) {
            throw ProcessingError("colorize expects amount:color");
        }
        return ops::colorize(img, parseColor(parts[1]), percentOrUnit(parts[0], 1.0));
    }
    if (name == "tint") {
        auto parts = splitEffectArgs(value);
        if (parts.size() < 2) {
            throw ProcessingError("tint expects amount:color");
        }
        return ops::tint(img, parseColor(parts[1]), percentOrUnit(parts[0], 1.0));
    }
    if (name == "duotone") {
        auto parts = splitEffectArgs(value);
        if (parts.size() < 2) {
            throw ProcessingError("duotone expects shadow:highlight");
        }
        return ops::duotone(img, parseColor(parts[0]), parseColor(parts[1]));
    }
    throw ProcessingError("unknown effect: " + name);
}

} // namespace transform
} // namespace voxire
