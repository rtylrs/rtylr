#pragma once

#include <string>

namespace voxire {
namespace video {

// probe returns ffprobe's JSON description of the input (format + streams),
// used by the `video probe` subcommand. Throws ProcessingError on failure.
std::string probe(const std::string& inputPath);

} // namespace video
} // namespace voxire
