#include "image_processor.hpp"
#include "video_processor.hpp"
#include <vips/vips8>
#include <iostream>
#include <string>
#include <vector>
#include <unordered_map>
#include <stdexcept>

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
        if (token.rfind("--", 0) == 0 && i + 1 < argc) {
            args.flags[token.substr(2)] = argv[++i];
        }
    }
    return args;
}

int getInt(const Args& args, const std::string& key, int fallback = 0) {
    auto it = args.flags.find(key);
    if (it == args.flags.end()) return fallback;
    return std::stoi(it->second);
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
// Entry point
// ---------------------------------------------------------------------------

int main(int argc, char* argv[]) {
    if (argc < 2) {
        std::cerr << "Usage: media-engine <subcommand> [options]\n"
                  << "Subcommands:\n"
                  << "  resize     --input <path> --output <path> --width N --height N\n"
                  << "             --format <webp|jpeg|png|avif> --quality N --fit <cover|contain|fill>\n"
                  << "  thumbnail  --input <path> --output <path> --time N --width N --height N\n";
        return 1;
    }

    if (VIPS_INIT(argv[0])) {
        vips_error_exit("vips_init");
    }

    int exitCode = 0;
    try {
        Args args = parseArgs(argc, argv);

        if (args.subcommand == "resize") {
            cmdResize(args);
        } else if (args.subcommand == "thumbnail") {
            cmdThumbnail(args);
        } else {
            std::cerr << "Unknown subcommand: " << args.subcommand << "\n";
            exitCode = 1;
        }
    } catch (const std::exception& e) {
        std::cerr << "error: " << e.what() << "\n";
        exitCode = 1;
    }

    vips_shutdown();
    return exitCode;
}
