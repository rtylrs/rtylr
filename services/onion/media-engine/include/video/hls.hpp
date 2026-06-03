#pragma once

#include <string>

namespace voxire {
namespace video {

// packageHLS transcodes the input into an HLS VOD stream: an .m3u8 playlist at
// playlistPath plus .ts segments alongside it. segmentSeconds sets the target
// segment length. Throws ProcessingError on failure.
void packageHLS(const std::string& inputPath, const std::string& playlistPath,
                int segmentSeconds);

} // namespace video
} // namespace voxire
