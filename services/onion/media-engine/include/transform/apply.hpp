#pragma once

#include <string>
#include <vips/vips8>

namespace voxire {
namespace transform {

// applyChain applies a full Cloudinary-style transformation chain (steps
// separated by '/', params by ',') to img, in order. Pixel-affecting params
// are handled here; output format/quality (f_/q_) are encode-time and ignored.
// Throws ProcessingError on an unknown directive.
vips::VImage applyChain(vips::VImage img, const std::string& chain);

} // namespace transform
} // namespace voxire
