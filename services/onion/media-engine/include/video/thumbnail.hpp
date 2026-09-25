#pragma once

#include <string>

namespace voxire {
namespace video {

// extractThumbnail writes a single still frame from the video at `timeSecs`
// into outputPath, scaled to width x height (0 in a dimension keeps the aspect
// ratio). The output format is inferred from the output extension.
void extractThumbnail(const std::string& inputPath, const std::string& outputPath,
                      double timeSecs, int width, int height);

} // namespace video
} // namespace voxire
