#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// Split separates an image into its color bands and an optional alpha band so
// color operations (brightness, tint, etc.) can run without disturbing
// transparency, then be recombined.
struct Split {
    vips::VImage color;
    vips::VImage alpha;
    bool hasAlpha = false;
};

// splitAlpha peels the trailing alpha band off img when present.
inline Split splitAlpha(vips::VImage img) {
    Split s;
    if (img.has_alpha()) {
        const int n = img.bands();
        s.alpha = img.extract_band(n - 1);
        s.color = img.extract_band(0, vips::VImage::option()->set("n", n - 1));
        s.hasAlpha = true;
    } else {
        s.color = img;
    }
    return s;
}

// joinAlpha reattaches the alpha band that splitAlpha removed (no-op when the
// source had none).
inline vips::VImage joinAlpha(vips::VImage color, const Split& s) {
    if (s.hasAlpha) {
        return color.bandjoin(s.alpha);
    }
    return color;
}

} // namespace ops
} // namespace voxire
