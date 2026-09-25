#pragma once

#include <string>

namespace voxire {
namespace video {

// extractAudio writes the input's audio track to outputPath, copying the codec
// without re-encoding (so the output container must match the source audio
// codec — e.g. .m4a/.aac for AAC). Throws ProcessingError on failure.
void extractAudio(const std::string& inputPath, const std::string& outputPath);

} // namespace video
} // namespace voxire
