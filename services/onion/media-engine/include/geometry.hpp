#pragma once

#include "common.hpp"
#include <string>
#include <cstdlib>

namespace voxire {

// Size is a width/height pair in pixels. A zero dimension means "unset" and is
// resolved against the source aspect ratio by the caller.
struct Size {
    int width  = 0;
    int height = 0;

    bool empty() const { return width == 0 && height == 0; }
    long area() const { return static_cast<long>(width) * static_cast<long>(height); }
};

// Box is an axis-aligned rectangle: an (x,y) top-left origin plus a Size.
// (Named Box rather than Rect because libvips defines a `#define Rect VipsRect`
// compatibility macro that textually clobbers the token in any file that
// includes <vips/vips8>.)
struct Box {
    int x = 0;
    int y = 0;
    int width  = 0;
    int height = 0;

    int right()  const { return x + width; }
    int bottom() const { return y + height; }
    bool empty() const { return width <= 0 || height <= 0; }
};

// parseSize parses "WxH", "Wx", or "xH" (e.g. "256x256", "256x", "x256").
// A missing component is left as 0. Throws ProcessingError on a malformed value.
inline Size parseSize(const std::string& s) {
    auto pos = s.find('x');
    if (pos == std::string::npos) {
        throw ProcessingError("invalid size (expected WxH): " + s);
    }
    Size out;
    const std::string w = s.substr(0, pos);
    const std::string h = s.substr(pos + 1);
    if (!w.empty()) out.width = std::atoi(w.c_str());
    if (!h.empty()) out.height = std::atoi(h.c_str());
    if (out.width < 0 || out.height < 0) {
        throw ProcessingError("size dimensions must not be negative: " + s);
    }
    return out;
}

} // namespace voxire
