#include "video/poster.hpp"
#include "video/thumbnail.hpp"

namespace voxire {
namespace video {

void extractPoster(const std::string& inputPath, const std::string& outputPath,
                   int width, int height) {
    // The poster is simply the opening frame.
    extractThumbnail(inputPath, outputPath, 0.0, width, height);
}

} // namespace video
} // namespace voxire
