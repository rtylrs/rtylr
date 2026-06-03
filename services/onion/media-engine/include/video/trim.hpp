#pragma once

#include <string>

namespace voxire {
namespace video {

// trim cuts the segment [startSecs, startSecs+durationSecs) from the input,
// copying streams without re-encoding for speed. Throws ProcessingError on
// failure or invalid bounds.
void trim(const std::string& inputPath, const std::string& outputPath,
          double startSecs, double durationSecs);

} // namespace video
} // namespace voxire
