#include "video/scale.hpp"
#include "video/util.hpp"
#include "common.hpp"

namespace voxire {
namespace video {

void scaleVideo(const std::string& inputPath, const std::string& outputPath,
                int width, int height) {
    if (width <= 0 && height <= 0) {
        throw ProcessingError("video scale needs a positive width or height");
    }
    runFFmpeg({
        "-i", inputPath,
        "-vf", scaleFilter(width, height),
        "-c:a", "copy",
        outputPath,
    });
}

} // namespace video
} // namespace voxire
