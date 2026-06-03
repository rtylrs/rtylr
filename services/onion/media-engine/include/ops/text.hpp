#pragma once

#include "../color.hpp"
#include <string>
#include <vips/vips8>

namespace voxire {
namespace ops {

// renderText rasterizes `text` in the given Pango font (e.g. "sans 48",
// "DejaVu Bold 32") wrapped to at most `maxWidth` pixels (0 = no wrap), colored
// with `color`. The result is an RGBA image sized to the text, ready to be
// composited as a text overlay.
vips::VImage renderText(const std::string& text, const std::string& font,
                        int maxWidth, const Color& color);

} // namespace ops
} // namespace voxire
