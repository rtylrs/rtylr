#include "video/fps.hpp"
#include "video/util.hpp"
#include "common.hpp"

namespace voxire {
namespace video {

void changeFps(const std::string& inputPath, const std::string& outputPath, int fps) {
    if (fps <= 0) {
        throw ProcessingError("fps must be greater than 0");
    }
    runFFmpeg({
        "-i", inputPath,
        "-vf", "fps=" + std::to_string(fps),
        "-c:a", "copy",
        outputPath,
    });
}

} // namespace video
} // namespace voxire
