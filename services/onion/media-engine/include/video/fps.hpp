#pragma once

#include <string>

namespace voxire {
namespace video {

// changeFps re-encodes the video to the target frame rate (dropping or
// duplicating frames as needed), copying the audio. fps must be > 0. Throws
// ProcessingError on failure.
void changeFps(const std::string& inputPath, const std::string& outputPath, int fps);

} // namespace video
} // namespace voxire
