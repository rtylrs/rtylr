#include "image_processor.hpp"
#include "video_processor.hpp"
#include "transform/apply.hpp"
#include "transform/params.hpp"
#include "transform/presets.hpp"
#include "encode.hpp"
#include "analyze.hpp"
#include "ops/blurhash.hpp"
#include "ops/average_color.hpp"
#include "ops/phash.hpp"
#include "video/probe.hpp"
#include "video/thumbnail.hpp"
#include "video/poster.hpp"
#include "video/transcode.hpp"
#include "video/hls.hpp"
#include "video/gif.hpp"
#include "video/trim.hpp"
#include "video/scale.hpp"
#include "video/mute.hpp"
#include "video/fps.hpp"
#include "video/audio.hpp"
#include <vips/vips8>
#include <iostream>
#include <string>
#include <vector>
#include <unordered_map>
#include <stdexcept>
#include <cstdio>

// MEDIA_ENGINE_VERSION is injected by CMake; fall back to "dev" for ad-hoc builds.
#ifndef MEDIA_ENGINE_VERSION
#define MEDIA_ENGINE_VERSION "dev"
#endif

// ---------------------------------------------------------------------------
// Minimal CLI argument parser
// ---------------------------------------------------------------------------

struct Args {
    std::string              subcommand;
    std::unordered_map<std::string, std::string> flags;
};

Args parseArgs(int argc, char* argv[]) {
    Args args;
    if (argc < 2) return args;

    args.subcommand = argv[1];

    for (int i = 2; i < argc; ++i) {
        std::string token = argv[i];
        if (token.rfind("--", 0) == 0) {
            if (i + 1 >= argc) {
                throw std::runtime_error("flag " + token + " requires a value");
            }
            args.flags[token.substr(2)] = argv[++i];
        }
    }
    return args;
}

int getInt(const Args& args, const std::string& key, int fallback = 0) {
    auto it = args.flags.find(key);
    if (it == args.flags.end()) return fallback;
    try {
        return std::stoi(it->second);
    } catch (const std::exception&) {
        throw std::runtime_error("invalid integer for --" + key + ": " + it->second);
    }
}

std::string getStr(const Args& args, const std::string& key, const std::string& fallback = "") {
    auto it = args.flags.find(key);
    if (it == args.flags.end()) return fallback;
    return it->second;
}

// ---------------------------------------------------------------------------
// Subcommand: resize
//
// Usage:
//   media-engine resize \
//     --input  input.jpg   \
//     --output output.webp \
//     --width  256         \
//     --height 256         \
//     --format webp        \
//     --quality 80         \
//     --fit cover
// ---------------------------------------------------------------------------

static void cmdResize(const Args& args) {
    voxire::ImageOptions opts;
    opts.inputPath  = getStr(args, "input");
    opts.outputPath = getStr(args, "output");
    opts.format     = getStr(args, "format", "webp");
    opts.width      = getInt(args, "width",   0);
    opts.height     = getInt(args, "height",  0);
    opts.quality    = getInt(args, "quality", 85);
    opts.fit        = voxire::parseFitMode(getStr(args, "fit", "cover"));

    if (opts.inputPath.empty() || opts.outputPath.empty()) {
        throw std::runtime_error("--input and --output are required");
    }

    const std::string& f = opts.format;
    if (f != "webp" && f != "jpeg" && f != "jpg" && f != "png" && f != "avif") {
        throw std::runtime_error("unsupported output format: " + f);
    }

    voxire::ImageProcessor processor;
    processor.process(opts);
}

// ---------------------------------------------------------------------------
// Subcommand: thumbnail (future)
// ---------------------------------------------------------------------------

static void cmdThumbnail(const Args& args) {
    voxire::VideoOptions opts;
    opts.inputPath        = getStr(args, "input");
    opts.outputPath       = getStr(args, "output");
    opts.thumbnailTimeSecs = getInt(args, "time", 1);
    opts.width            = getInt(args, "width",  0);
    opts.height           = getInt(args, "height", 0);

    voxire::VideoProcessor processor;
    processor.extractThumbnail(opts);
}

// ---------------------------------------------------------------------------
// Subcommand: transform
//
// Applies a Cloudinary-style transformation chain, e.g.
//   media-engine transform --input in.jpg --output out.webp \
//     --chain "w_400,h_400,c_fill,g_auto/e_blur:300" --format auto
// ---------------------------------------------------------------------------

static void cmdTransform(const Args& args) {
    const std::string in = getStr(args, "input");
    const std::string out = getStr(args, "output");
    std::string chain = getStr(args, "chain");
    if (in.empty() || out.empty()) {
        throw std::runtime_error("--input and --output are required");
    }
    // A named preset expands to a chain; an explicit --chain is appended after
    // it so callers can layer extra directives on top.
    const std::string preset = getStr(args, "preset");
    if (!preset.empty()) {
        const std::string base = voxire::transform::presetChain(preset);
        if (base.empty()) {
            throw std::runtime_error("unknown preset: " + preset);
        }
        chain = chain.empty() ? base : base + "/" + chain;
    }

    vips::VImage img = vips::VImage::new_from_file(
        in.c_str(), vips::VImage::option()->set("access", VIPS_ACCESS_SEQUENTIAL));
    img = img.autorot();
    img = voxire::transform::applyChain(img, chain);

    voxire::EncodeOptions enc;
    // Honor f_/q_ embedded in the chain, then let explicit flags override.
    for (const auto& step : voxire::transform::splitSteps(chain)) {
        auto p = voxire::transform::parseParams(step);
        if (p.count("f")) enc.format = voxire::parseOutputFormat(p["f"]);
        if (p.count("q") && p["q"].rfind("auto", 0) != 0) enc.quality = std::stoi(p["q"]);
    }
    const std::string fmtFlag = getStr(args, "format");
    if (!fmtFlag.empty()) enc.format = voxire::parseOutputFormat(fmtFlag);
    const int qFlag = getInt(args, "quality", 0);
    if (qFlag > 0) enc.quality = qFlag;

    voxire::encodeToFile(img, out, enc);
}

// ---------------------------------------------------------------------------
// Subcommand: info — print image metadata as JSON
// ---------------------------------------------------------------------------

static void cmdInfo(const Args& args) {
    const std::string in = getStr(args, "input");
    if (in.empty()) {
        throw std::runtime_error("--input is required");
    }
    vips::VImage img = vips::VImage::new_from_file(in.c_str());
    std::cout << voxire::infoToJSON(voxire::probeInfo(img)) << "\n";
}

// ---------------------------------------------------------------------------
// Subcommand: analyze — emit a metric for the input
//   media-engine analyze --input x.jpg --metric blurhash|average-color|phash
// ---------------------------------------------------------------------------

static void cmdAnalyze(const Args& args) {
    const std::string in = getStr(args, "input");
    if (in.empty()) {
        throw std::runtime_error("--input is required");
    }
    const std::string metric = getStr(args, "metric", "info");
    vips::VImage img = vips::VImage::new_from_file(
        in.c_str(), vips::VImage::option()->set("access", VIPS_ACCESS_SEQUENTIAL));

    if (metric == "info") {
        std::cout << voxire::infoToJSON(voxire::probeInfo(img)) << "\n";
    } else if (metric == "blurhash") {
        const int xc = getInt(args, "x", 4);
        const int yc = getInt(args, "y", 3);
        std::cout << voxire::ops::blurHash(img, xc, yc) << "\n";
    } else if (metric == "average-color") {
        const voxire::Color c = voxire::ops::averageColor(img);
        std::printf("#%02x%02x%02x\n", c.r, c.g, c.b);
    } else if (metric == "phash") {
        std::printf("%016llx\n",
            static_cast<unsigned long long>(voxire::ops::averageHash(img)));
    } else {
        throw std::runtime_error("unknown metric: " + metric);
    }
}

// ---------------------------------------------------------------------------
// Subcommand: video — dispatch the FFmpeg-backed operations
//   media-engine video --op <probe|thumbnail|poster|transcode|hls|gif|trim|
//                            scale|mute> --input <path> [--output <path>] ...
// ---------------------------------------------------------------------------

static void cmdVideo(const Args& args) {
    const std::string op = getStr(args, "op");
    const std::string in = getStr(args, "input");
    const std::string out = getStr(args, "output");
    if (op.empty() || in.empty()) {
        throw std::runtime_error("video requires --op and --input");
    }
    namespace v = voxire::video;
    const int w = getInt(args, "width", 0);
    const int h = getInt(args, "height", 0);

    if (op == "probe") {
        std::cout << v::probe(in) << "\n";
        return;
    }
    if (out.empty()) {
        throw std::runtime_error("video --op " + op + " requires --output");
    }
    if (op == "thumbnail") {
        v::extractThumbnail(in, out, getInt(args, "time", 1), w, h);
    } else if (op == "poster") {
        v::extractPoster(in, out, w, h);
    } else if (op == "transcode") {
        v::transcode(in, out, v::parseVideoCodec(getStr(args, "codec", "h264")),
                     getInt(args, "crf", 23), w, h);
    } else if (op == "hls") {
        v::packageHLS(in, out, getInt(args, "segment", 6));
    } else if (op == "gif") {
        v::makeGifPreview(in, out, getInt(args, "fps", 12), w > 0 ? w : 480);
    } else if (op == "trim") {
        v::trim(in, out, getInt(args, "start", 0), getInt(args, "duration", 5));
    } else if (op == "scale") {
        v::scaleVideo(in, out, w, h);
    } else if (op == "mute") {
        v::muteVideo(in, out);
    } else if (op == "fps") {
        v::changeFps(in, out, getInt(args, "fps", 30));
    } else if (op == "audio") {
        v::extractAudio(in, out);
    } else {
        throw std::runtime_error("unknown video op: " + op);
    }
}

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

static void printUsage(std::ostream& os) {
    os << "media-engine " << MEDIA_ENGINE_VERSION << "\n"
       << "Usage: media-engine <subcommand> [options]\n"
       << "Subcommands:\n"
       << "  resize     --input <path> --output <path> --width N --height N\n"
       << "             --format <webp|jpeg|png|avif> --quality N\n"
       << "             --fit <cover|contain|fill|inside|outside>\n"
       << "  transform  --input <path> --output <path> --chain <cloudinary-chain>\n"
       << "             [--preset <thumb|avatar|card|hero|og|blur>]\n"
       << "             [--format <auto|webp|jpeg|png|avif|gif|tiff>] [--quality N]\n"
       << "  info       --input <path>   (print image metadata as JSON)\n"
       << "  analyze    --input <path> --metric <info|blurhash|average-color|phash>\n"
       << "  video      --op <probe|thumbnail|poster|transcode|hls|gif|trim|scale|\n"
       << "                    mute|fps|audio>\n"
       << "             --input <path> [--output <path>] [op-specific flags]\n"
       << "  thumbnail  --input <path> --output <path> --time N --width N --height N\n"
       << "  help       show this message\n"
       << "  version    print the version and exit\n";
}

int main(int argc, char* argv[]) {
    if (argc < 2) {
        printUsage(std::cerr);
        return 1;
    }

    // Handle help/version before initialising libvips — they need nothing else.
    std::string sub = argv[1];
    if (sub == "help" || sub == "--help" || sub == "-h") {
        printUsage(std::cout);
        return 0;
    }
    if (sub == "version" || sub == "--version" || sub == "-v") {
        std::cout << "media-engine " << MEDIA_ENGINE_VERSION << "\n";
        return 0;
    }

    if (VIPS_INIT(argv[0])) {
        vips_error_exit("vips_init");
    }

    int exitCode = 0;
    try {
        Args args = parseArgs(argc, argv);

        if (args.subcommand == "resize") {
            cmdResize(args);
        } else if (args.subcommand == "transform") {
            cmdTransform(args);
        } else if (args.subcommand == "info") {
            cmdInfo(args);
        } else if (args.subcommand == "analyze") {
            cmdAnalyze(args);
        } else if (args.subcommand == "video") {
            cmdVideo(args);
        } else if (args.subcommand == "thumbnail") {
            cmdThumbnail(args);
        } else {
            std::cerr << "Unknown subcommand: " << args.subcommand << "\n\n";
            printUsage(std::cerr);
            exitCode = 1;
        }
    } catch (const std::exception& e) {
        std::cerr << "error: " << e.what() << "\n";
        exitCode = 1;
    }

    vips_shutdown();
    return exitCode;
}
