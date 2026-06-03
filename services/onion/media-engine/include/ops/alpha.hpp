#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// addAlpha attaches a uniform alpha band with the given value (0-255) when the
// image has none; a no-op if it already has alpha.
vips::VImage addAlpha(vips::VImage img, int value);

// removeAlpha drops the alpha band, returning the color bands only; a no-op if
// the image has no alpha.
vips::VImage removeAlpha(vips::VImage img);

} // namespace ops
} // namespace voxire
