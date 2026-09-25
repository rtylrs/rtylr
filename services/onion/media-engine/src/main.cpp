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
#include <vips/error.h>
#include <iostream>
#include <iomanip>
#include <sstream>
#include <string>
#include <vector>
#include <unordered_map>
#include <stdexcept>
#include <cstdio>
#include <cstdint>
#include <cstring>

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

Args parseArgs(const std::vector<std::string>& tokens) {
    Args args;
    if (tokens.empty()) return args;

    args.subcommand = tokens[0];

    for (size_t i = 1; i < tokens.size(); ++i) {
        const std::string& token = tokens[i];
        if (token.rfind("--", 0) == 0) {
            if (i + 1 >= tokens.size()) {
                throw std::runtime_error("flag " + token + " requires a value");
            }
            args.flags[token.substr(2)] = tokens[++i];
        }
    }
    return args;
}

Args parseArgs(int argc, char* argv[]) {
    return parseArgs(std::vector<std::string>(argv + 1, argv + argc));
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

    vips::VImage img;
    // Fold the leading geometry step into the decode (shrink-on-load) when
    // the chain shape allows it; otherwise decode fully and apply steps.
    auto fast = voxire::transform::tryApplyChainFromFile(in, chain);
    if (fast) {
        img = std::move(*fast);
    } else {
        img = vips::VImage::new_from_file(
            in.c_str(), vips::VImage::option()->set("access", VIPS_ACCESS_SEQUENTIAL));
        img = img.autorot();
        img = voxire::transform::applyChain(img, chain);
    }

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

    // Encoder effort overrides (see EncodeOptions for defaults and tradeoffs).
    const int effortFlag = getInt(args, "effort", 0);
    if (effortFlag > 0) enc.webpEffort = effortFlag;
    const int avifSpeedFlag = getInt(args, "avif-speed", -1);
    if (avifSpeedFlag >= 0) enc.avifSpeed = avifSpeedFlag;
    const int pngCompFlag = getInt(args, "png-compression", 0);
    if (pngCompFlag > 0) enc.pngCompression = pngCompFlag;

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
        // std::cout rather than printf so resident mode can capture output.
        std::cout << '#' << std::hex << std::setfill('0')
                  << std::setw(2) << c.r << std::setw(2) << c.g
                  << std::setw(2) << c.b << std::dec << '\n';
    } else if (metric == "phash") {
        std::cout << std::hex << std::setfill('0') << std::setw(16)
                  << static_cast<unsigned long long>(voxire::ops::averageHash(img))
                  << std::dec << '\n';
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
        << "  --serve    resident mode: length-prefixed jobs on stdin/stdout\n"
        << "  help       show this message\n"
        << "  version    print the version and exit\n";
}

// ---------------------------------------------------------------------------
// Command dispatch
// ---------------------------------------------------------------------------

// Runs one parsed command vector (argv-style, without the binary name).
// Returns the process exit code for that command.
static int runCommand(const std::vector<std::string>& argVector) {
    try {
        Args args = parseArgs(argVector);

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
            return 1;
        }
        return 0;
    } catch (const std::exception& e) {
        std::cerr << "error: " << e.what() << "\n";
        return 1;
    }
}

// ---------------------------------------------------------------------------
// Resident mode (--serve)
//
// Amortises process startup (~60ms of fork/exec, dynamic linking and
// VIPS_INIT) across many jobs. Callers speak a length-prefixed framing on
// stdin/stdout:
//
//   request  : u32le payloadLen, payload = argv tokens joined with '\0'
//   response : u32le exitCode, u32le msgLen, msg bytes (= command stdout)
//
// A zero-length request or EOF terminates the loop. Image data still flows
// through --input/--output files, so no pixel serialisation is involved.
// ---------------------------------------------------------------------------

static bool readExact(unsigned char* buf, size_t n) {
    size_t got = 0;
    while (got < n) {
        const size_t r = std::fread(buf + got, 1, n - got, stdin);
        if (r == 0) return false;
        got += r;
    }
    return true;
}

static bool writeExact(const unsigned char* buf, size_t n) {
    size_t put = 0;
    while (put < n) {
        const size_t w = std::fwrite(buf + put, 1, n - put, stdout);
        if (w == 0) return false;
        put += w;
    }
    return true;
}

static int serveLoop() {
    while (true) {
        unsigned char lenBuf[4];
        if (!readExact(lenBuf, 4)) return 0; // EOF: clean shutdown
        const uint32_t payloadLen =
            static_cast<uint32_t>(lenBuf[0]) |
            (static_cast<uint32_t>(lenBuf[1]) << 8) |
            (static_cast<uint32_t>(lenBuf[2]) << 16) |
            (static_cast<uint32_t>(lenBuf[3]) << 24);
        if (payloadLen == 0) return 0;       // explicit quit

        if (payloadLen > 64 * 1024) {
            std::cerr << "serve: rejecting oversized frame (" << payloadLen << ")\n";
            return 1;
        }

        std::vector<unsigned char> payload(payloadLen);
        if (!readExact(payload.data(), payload.size())) return 0;

        // Split NUL-separated tokens into an argv-style vector. The trailing
        // NUL is optional; never read past the frame boundary.
        std::vector<std::string> tokens;
        const char* p = reinterpret_cast<const char*>(payload.data());
        const char* end = p + payloadLen;
        while (p < end) {
            const void* from = p;
            const size_t len = strnlen(static_cast<const char*>(from),
                                       static_cast<size_t>(end - p));
            tokens.emplace_back(p, len);
            p += len + 1;
        }

        // Capture command stdout (info/analyze output) so it can be framed
        // back to the caller instead of leaking into the framing channel.
        std::ostringstream captured;
        std::streambuf* const original = std::cout.rdbuf(captured.rdbuf());
        const int code = runCommand(tokens);
        std::cout.rdbuf(original); // restore
        vips_error_clear(); // don't let one failure poison the next job

        const std::string msg = captured.str();
        const uint32_t status = static_cast<uint32_t>(code);
        const uint32_t msgLen = static_cast<uint32_t>(msg.size());
        unsigned char head[8] = {
            static_cast<unsigned char>(status),
            static_cast<unsigned char>(status >> 8),
            static_cast<unsigned char>(status >> 16),
            static_cast<unsigned char>(status >> 24),
            static_cast<unsigned char>(msgLen),
            static_cast<unsigned char>(msgLen >> 8),
            static_cast<unsigned char>(msgLen >> 16),
            static_cast<unsigned char>(msgLen >> 24),
        };
        if (!writeExact(head, 8) ||
            (!msg.empty() && !writeExact(
                 reinterpret_cast<const unsigned char*>(msg.data()), msg.size()))) {
            return 1;
        }
        std::fflush(stdout);
    }
}

int main(int argc, char* argv[]) {
    if (argc >= 2) {
        const std::string sub = argv[1];

        // Help/version need nothing but themselves.
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
        if (sub == "--serve") {
            exitCode = serveLoop();
        } else {
            std::vector<std::string> argVector(argv + 1, argv + argc);
            exitCode = runCommand(argVector);
        }

        vips_shutdown();
        return exitCode;
    }

    printUsage(std::cerr);
    return 1;
}
