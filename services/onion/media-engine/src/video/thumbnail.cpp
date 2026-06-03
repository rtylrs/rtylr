#include "video/thumbnail.hpp"
#include "video/util.hpp"
#include "common.hpp"

namespace voxire {
namespace video {

void extractThumbnail(const std::string& inputPath, const std::string& outputPath,
                      double timeSecs, int width, int height) {
    if (timeSecs < 0) {
        throw ProcessingError("thumbnail time must not be negative");
    }
    // -ss before -i seeks fast (keyframe-accurate); one frame; scaled.
    runFFmpeg({
        "-ss", std::to_string(timeSecs),
        "-i", inputPath,
        "-frames:v", "1",
        "-vf", scaleFilter(width, height),
        outputPath,
    });
}

} // namespace video
} // namespace voxire
