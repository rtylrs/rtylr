#include "video/hls.hpp"
#include "video/util.hpp"
#include "common.hpp"
#include <string>

namespace voxire {
namespace video {

void packageHLS(const std::string& inputPath, const std::string& playlistPath,
                int segmentSeconds) {
    if (segmentSeconds <= 0) {
        throw ProcessingError("HLS segment length must be greater than 0");
    }
    // Derive the segment filename pattern from the playlist path so segments
    // land next to the .m3u8.
    std::string base = playlistPath;
    const auto dot = base.find_last_of('.');
    if (dot != std::string::npos) {
        base = base.substr(0, dot);
    }
    const std::string segmentPattern = base + "_%03d.ts";

    runFFmpeg({
        "-i", inputPath,
        "-c:v", "libx264", "-crf", "23", "-preset", "medium",
        "-pix_fmt", "yuv420p",
        "-c:a", "aac", "-b:a", "128k",
        "-hls_time", std::to_string(segmentSeconds),
        "-hls_playlist_type", "vod",
        "-hls_segment_filename", segmentPattern,
        playlistPath,
    });
}

} // namespace video
} // namespace voxire
