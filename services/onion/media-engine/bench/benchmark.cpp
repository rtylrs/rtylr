// Benchmark harness for media-engine.
//
// Self-contained: generates synthetic test images on first run, then times the
// main processing paths with warmup + repeated runs (median/min reported).
//
// Usage:
//   media-engine-bench [--iters N] [--warmup N] [--filter substr] [--csv]
//
// Interpreting results: run before and after a change on the same quiet
// machine and compare medians. Min is useful to spot scheduling noise.

#include "image_processor.hpp"
#include "transform/apply.hpp"
#include "transform/params.hpp"
#include "encode.hpp"
#include "analyze.hpp"
#include "ops/blurhash.hpp"
#include "ops/average_color.hpp"
#include "ops/phash.hpp"

#include <vips/vips8>
#include <algorithm>
#include <chrono>
#include <cstdlib>
#include <cstring>
#include <filesystem>
#include <fstream>
#include <functional>
#include <iomanip>
#include <iostream>
#include <string>
#include <vector>

namespace {

namespace fs = std::filesystem;

const char* kVersion = "1";

fs::path benchDir() {
    const char* env = std::getenv("MEDIA_ENGINE_BENCH_DIR");
    if (env && *env) return fs::path(env);
    return fs::temp_directory_path() / "media-engine-bench";
}

// Deterministic pseudo-noise so generated images are identical across runs.
double noise(int x, int y, int c) {
    const unsigned h = static_cast<unsigned>(x * 374761393u + y * 668265263u + c * 2246822519u);
    return ((h ^ (h >> 13)) * 1274126177u >> 16 & 1023) / 1023.0;
}

vips::VImage makeSynthetic(int width, int height, bool alpha) {
    // Smooth gradient + sinusoidal detail + mild noise: exercises resampling,
    // colour transforms and encoders more faithfully than a flat fill.
    // Pixels are written into a plain buffer so fixtures are byte-identical
    // across runs and never depend on lazy-pipeline evaluation order.
    const int bands = alpha ? 4 : 3;
    std::vector<unsigned char> px(static_cast<size_t>(width) * height * bands);
    for (int y = 0; y < height; ++y) {
        for (int x = 0; x < width; ++x) {
            const double u = static_cast<double>(x) / width;
            const double v = static_cast<double>(y) / height;
            unsigned char* p = px.data() + (static_cast<size_t>(y) * width + x) * bands;
            auto clamp8 = [](double c) -> unsigned char {
                int i = static_cast<int>(c * 255.0 + 0.5);
                return static_cast<unsigned char>(i < 0 ? 0 : (i > 255 ? 255 : i));
            };
            p[0] = clamp8(0.5 * (u + v) + 0.15 * noise(x, y, 0));
            p[1] = clamp8(0.3 * u + 0.2 * std::sin(v * 12.0) + 0.4 + 0.1 * noise(x, y, 1));
            p[2] = clamp8(0.15 * std::cos(u * 10.0) + 0.7 * v + 0.2 + 0.1 * noise(x, y, 2));
            if (alpha) p[3] = clamp8(0.2 + 0.6 * u);
        }
    }
    // new_from_memory_copy duplicates the buffer into vips-owned memory, so
    // the local pixel vector can go out of scope safely.
    return vips::VImage::new_from_memory_copy(
        px.data(), static_cast<size_t>(px.size()), width, height, bands,
        VIPS_FORMAT_UCHAR);
}

// Creates the fixture set once and returns their paths.
struct Fixtures {
    fs::path photo4k;   // 3840x2160 opaque, jpeg — decode/resample heavy path
    fs::path graphic;   // 1600x900 with alpha, png — alpha-aware paths
    fs::path small;     // 320x240 webp — cheap-path overhead measurement
};

Fixtures ensureFixtures() {
    Fixtures f;
    fs::create_directories(benchDir());
    f.photo4k = benchDir() / "photo4k.jpg";
    f.graphic = benchDir() / "graphic.png";
    f.small   = benchDir() / "small.webp";

    if (!fs::exists(f.photo4k)) {
        makeSynthetic(3840, 2160, false).jpegsave(f.photo4k.c_str(),
            vips::VImage::option()->set("Q", 85));
    }
    if (!fs::exists(f.graphic)) {
        makeSynthetic(1600, 900, true).pngsave(f.graphic.c_str());
    }
    if (!fs::exists(f.small)) {
        makeSynthetic(320, 240, false).webpsave(f.small.c_str(),
            vips::VImage::option()->set("Q", 80));
    }
    return f;
}

struct Result {
    std::string name;
    int iters = 0;
    double medianMs = 0;
    double minMs = 0;
};

class Bench {
public:
    Bench(int warmup, int iters, const std::string& filter, bool csv)
        : warmup_(warmup), iters_(iters), filter_(filter), csv_(csv) {}

    void run(const std::string& name, const std::function<void()>& fn) {
        if (!filter_.empty() && name.find(filter_) == std::string::npos) return;
        for (int i = 0; i < warmup_; ++i) fn();
        std::vector<double> samples;
        samples.reserve(static_cast<size_t>(iters_));
        for (int i = 0; i < iters_; ++i) {
            const auto t0 = std::chrono::steady_clock::now();
            fn();
            const auto t1 = std::chrono::steady_clock::now();
            samples.push_back(std::chrono::duration<double, std::milli>(t1 - t0).count());
        }
        std::sort(samples.begin(), samples.end());
        Result r{name, iters_, samples[samples.size() / 2], samples.front()};
        results_.push_back(r);
        print(r);
    }

    void print(const Result& r) const {
        if (csv_) {
            std::cout << r.name << ',' << r.iters << ','
                      << std::fixed << std::setprecision(3) << r.medianMs << ','
                      << r.minMs << '\n';
        } else {
            std::cout << std::left  << std::setw(34) << r.name
                      << std::right << std::setw(10) << std::fixed << std::setprecision(2) << r.medianMs
                      << std::setw(10) << r.minMs << " ms\n";
        }
    }

    void header() const {
        if (!csv_) {
            std::cout << std::left << std::setw(34) << "benchmark"
                      << std::right << std::setw(10) << "median" << std::setw(14) << "min\n"
                      << std::string(58, '-') << '\n';
        }
    }

    const std::vector<Result>& results() const { return results_; }

private:
    int warmup_;
    int iters_;
    std::string filter_;
    bool csv_;
    std::vector<Result> results_;
};

void runAll(Bench& b, const Fixtures& fx) {
    const fs::path out = benchDir() / "out.bin";

    // Decode + resample + encode pipelines (the production hot paths).
    b.run("resize_4k_jpg_to_800_webp", [&] {
        voxire::ImageProcessor p;
        p.process({fx.photo4k.string(), out.string(), "webp", 800, 800, 85, voxire::FitMode::Cover});
    });
    b.run("resize_4k_jpg_to_1920_jpeg", [&] {
        vips::VImage img = vips::VImage::new_from_file(fx.photo4k.c_str(),
            vips::VImage::option()->set("access", VIPS_ACCESS_SEQUENTIAL));
        img = img.autorot();
        voxire::EncodeOptions e;
        e.format = voxire::OutputFormat::JPEG;
        e.quality = 82;
        voxire::encodeToFile(img.thumbnail_image(1920,
            vips::VImage::option()->set("size", VIPS_SIZE_DOWN)), out.string(), e);
    });
    b.run("resize_graphic_png_to_400_webp", [&] {
        vips::VImage img = vips::VImage::new_from_file(fx.graphic.c_str(),
            vips::VImage::option()->set("access", VIPS_ACCESS_SEQUENTIAL));
        img = img.autorot();
        voxire::EncodeOptions e;
        e.format = voxire::OutputFormat::WebP;
        e.quality = 85;
        voxire::encodeToFile(img.thumbnail_image(400,
            vips::VImage::option()->set("size", VIPS_SIZE_DOWN)), out.string(), e);
    });

    // Cloudinary-style transform chains end-to-end.
    b.run("transform_chain_fill_sharpen_webp", [&] {
        vips::VImage img = vips::VImage::new_from_file(fx.photo4k.c_str(),
            vips::VImage::option()->set("access", VIPS_ACCESS_SEQUENTIAL));
        img = img.autorot();
        img = voxire::transform::applyChain(img, "w_800,h_600,c_fill/e_sharpen:1.5");
        voxire::EncodeOptions e;
        e.format = voxire::OutputFormat::WebP;
        voxire::encodeToFile(img, out.string(), e);
    });
    // Same chain via the shrink-on-load fast path (first step at decode time).
    b.run("transform_sol_fill_sharpen_webp", [&] {
        auto img = voxire::transform::tryApplyChainFromFile(
            fx.photo4k.string(), "w_800,h_600,c_fill/e_sharpen:1.5");
        if (!img) throw std::runtime_error("fast path unexpectedly unavailable");
        voxire::EncodeOptions e;
        e.format = voxire::OutputFormat::WebP;
        voxire::encodeToFile(*img, out.string(), e);
    });
    // Centre-crop variants: shrink-on-load shines when no attention analysis
    // needs substantial pixels.
    b.run("transform_chain_centre_webp", [&] {
        vips::VImage img = vips::VImage::new_from_file(fx.photo4k.c_str(),
            vips::VImage::option()->set("access", VIPS_ACCESS_SEQUENTIAL));
        img = voxire::transform::applyChain(img, "w_800,h_600,c_fill,g_centre");
        voxire::EncodeOptions e;
        e.format = voxire::OutputFormat::WebP;
        voxire::encodeToFile(img, out.string(), e);
    });
    b.run("transform_sol_centre_webp", [&] {
        auto img = voxire::transform::tryApplyChainFromFile(
            fx.photo4k.string(), "w_800,h_600,c_fill,g_centre");
        if (!img) throw std::runtime_error("fast path unexpectedly unavailable");
        voxire::EncodeOptions e;
        e.format = voxire::OutputFormat::WebP;
        voxire::encodeToFile(*img, out.string(), e);
    });
    b.run("transform_chain_effects_png", [&] {
        vips::VImage img = vips::VImage::new_from_file(fx.graphic.c_str(),
            vips::VImage::option()->set("access", VIPS_ACCESS_SEQUENTIAL));
        img = voxire::transform::applyChain(img,
            "w_640,h_480,c_pad,b_white/e_grayscale/e_brightness:15/r_24");
        voxire::EncodeOptions e;
        e.format = voxire::OutputFormat::PNG;
        voxire::encodeToFile(img, out.string(), e);
    });

    // Analysis metrics (pure CPU over pixels). The image is fully decoded
    // into memory before timing so these measure metric computation, not I/O
    // (and so the shared instance can be evaluated repeatedly).
    {
        vips::VImage loaded = vips::VImage::new_from_file(fx.graphic.c_str())
            .copy_memory();
        b.run("analyze_blurhash_1600px", [&] {
            std::string h = voxire::ops::blurHash(loaded, 4, 3);
            if (h.empty()) throw std::runtime_error("empty blurhash");
        });
        b.run("analyze_phash_1600px", [&] {
            volatile uint64_t h = voxire::ops::averageHash(loaded);
            (void)h;
        });
        b.run("analyze_average_color_1600px", [&] {
            volatile voxire::Color c = voxire::ops::averageColor(loaded);
            (void)c;
        });
    }

    // Metadata-only path.
    b.run("info_json_4k", [&] {
        vips::VImage img = vips::VImage::new_from_file(fx.photo4k.c_str());
        std::string j = voxire::infoToJSON(voxire::probeInfo(img));
        if (j.empty()) throw std::runtime_error("empty info json");
    });

    fs::remove(out);
}

} // namespace

int main(int argc, char* argv[]) {
    int iters = 5, warmup = 2;
    std::string filter;
    bool csv = false;
    for (int i = 1; i < argc; ++i) {
        auto next = [&](const char* what) -> std::string {
            if (i + 1 >= argc) { std::cerr << "missing value for " << what << "\n"; std::exit(2); }
            return argv[++i];
        };
        std::string a = argv[i];
        if (a == "--iters")  iters = std::stoi(next("--iters"));
        else if (a == "--warmup") warmup = std::stoi(next("--warmup"));
        else if (a == "--filter") filter = next("--filter");
        else if (a == "--csv")    csv = true;
        else { std::cerr << "unknown flag: " << a << "\n"; return 2; }
    }
    if (iters < 1 || warmup < 0) { std::cerr << "bad iters/warmup\n"; return 2; }

    if (VIPS_INIT("media-engine-bench")) {
        vips_error_exit("vips_init");
    }

    std::cout << "media-engine-bench " << kVersion
              << " (vips " << VIPS_VERSION << ", " << vips_concurrency_get() << " threads)\n";
    try {
        Fixtures fx = ensureFixtures();
        Bench b(warmup, iters, filter, csv);
        b.header();
        runAll(b, fx);
    } catch (const std::exception& e) {
        std::cerr << "error: " << e.what() << "\n";
        vips_shutdown();
        return 1;
    }
    vips_shutdown();
    return 0;
}
