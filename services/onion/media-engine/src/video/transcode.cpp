#include "video/transcode.hpp"
#include "video/util.hpp"
#include "common.hpp"
#include <vector>

namespace voxire {
namespace video {

VideoCodec parseVideoCodec(const std::string& s) {
    if (s == "h264" || s == "mp4")  return VideoCodec::H264;
    if (s == "vp9"  || s == "webm") return VideoCodec::VP9;
    throw ProcessingError("unsupported video codec: " + s);
}

void transcode(const std::string& inputPath, const std::string& outputPath,
               VideoCodec codec, int crf, int width, int height) {
    if (crf < 0 || crf > 63) {
        throw ProcessingError("video crf must be between 0 and 63");
    }

    std::vector<std::string> args = {"-i", inputPath, "-vf", scaleFilter(width, height)};

    if (codec == VideoCodec::H264) {
        args.insert(args.end(), {
            "-c:v", "libx264", "-crf", std::to_string(crf), "-preset", "medium",
            "-pix_fmt", "yuv420p", "-c:a", "aac", "-b:a", "128k",
            "-movflags", "+faststart",
        });
    } else { // VP9 / WebM
        args.insert(args.end(), {
            "-c:v", "libvpx-vp9", "-crf", std::to_string(crf), "-b:v", "0",
            "-c:a", "libopus",
        });
    }
    args.push_back(outputPath);
    runFFmpeg(args);
}

} // namespace video
} // namespace voxire
