#pragma once

#include <optional>
#include <string>
#include <vips/vips8>

namespace voxire {
namespace transform {

// applyChain applies a full Cloudinary-style transformation chain (steps
// separated by '/', params by ',') to img, in order. Pixel-affecting params
// are handled here; output format/quality (f_/q_) are encode-time and ignored.
// Throws ProcessingError on an unknown directive.
vips::VImage applyChain(vips::VImage img, const std::string& chain);

// tryApplyChainFromFile folds the first geometry step into the file load via
// vips_thumbnail's shrink-on-load, then applies the remaining steps. Returns
// nullopt when the chain's first step is not expressible at decode time
// (callers should fall back to new_from_file + applyChain). Output pixels may
// differ microscopically from the slow path because decoders downscale before
// decompression; output dimensions are identical.
std::optional<vips::VImage> tryApplyChainFromFile(const std::string& path,
                                                  const std::string& chain);

} // namespace transform
} // namespace voxire
