#pragma once

#include <string>

namespace voxire {
namespace video {

// muteVideo strips the audio track, copying the video stream unchanged (fast,
// no re-encode). Throws ProcessingError on failure.
void muteVideo(const std::string& inputPath, const std::string& outputPath);

} // namespace video
} // namespace voxire
