#pragma once

#include <string>

namespace voxire {
namespace video {

// scaleVideo re-encodes the input scaled to width x height (0 in a dimension
// keeps the aspect ratio), copying the audio stream. Throws ProcessingError on
// failure.
void scaleVideo(const std::string& inputPath, const std::string& outputPath,
                int width, int height);

} // namespace video
} // namespace voxire
