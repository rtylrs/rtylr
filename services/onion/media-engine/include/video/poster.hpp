#pragma once

#include <string>

namespace voxire {
namespace video {

// extractPoster writes the first frame of the video as a still poster image,
// scaled to width x height (0 keeps aspect). Convenience over extractThumbnail
// at t=0.
void extractPoster(const std::string& inputPath, const std::string& outputPath,
                   int width, int height);

} // namespace video
} // namespace voxire
