#include "image_processor.hpp"
#include "check.hpp"
#include <vips/vips8>
#include <filesystem>
#include <iostream>
#include <stdexcept>
#include <string>
#include <unistd.h>

namespace fs = std::filesystem;

namespace {

// Runs one cover resize from a synthetic source and asserts the output has the
// exact target dimensions. reqW/reqH are what the caller asked for (0 means
// "derive it", as the CLI does); wantW/wantH are the expected output size.
void runCoverCase(const fs::path& dir, const std::string& name,
                  int srcW, int srcH, int reqW, int reqH, int wantW, int wantH) {
    const fs::path input = dir / (name + "_in.png");
    const fs::path output = dir / (name + "_out.png");

    vips::VImage::black(srcW, srcH, vips::VImage::option()->set("bands", 3))
        .pngsave(input.c_str());

    voxire::ImageProcessor processor;
    voxire::ImageOptions opts;
    opts.inputPath = input.string();
    opts.outputPath = output.string();
    opts.format = "png";
    opts.width = reqW;
    opts.height = reqH;
    opts.fit = voxire::FitMode::Cover;
    processor.process(opts);

    vips::VImage out = vips::VImage::new_from_file(output.c_str());
    CHECK_EQ(out.width(), wantW);
    CHECK_EQ(out.height(), wantH);

    fs::remove(input);
    fs::remove(output);
}

} // namespace

int main(int argc, char* argv[]) {
    if (VIPS_INIT(argv[0])) {
        vips_error_exit("vips_init");
    }

    const fs::path dir = fs::temp_directory_path();

    try {
        // Regression: the cover box was truncated to 750x562, vips fit inside
        // it and produced a 749x562 thumbnail, and the 750x500 crop then failed
        // with "bad extract area" (exit 1, no output file).
        runCoverCase(dir, "cover_4k", 4032, 3024, 750, 500, 750, 500);

        // Portrait source named in the same report.
        runCoverCase(dir, "cover_portrait", 1080, 1920, 100, 200, 100, 200);

        // Width-only request: the height is derived (1080 * 200 / 1920 = 112)
        // and the output used to come out 199x112.
        runCoverCase(dir, "cover_width_only", 1920, 1080, 200, 0, 200, 112);
    } catch (const std::exception& e) {
        std::cerr << "unexpected exception: " << e.what() << "\n";
        ++voxtest::failures;
    }

    vips_shutdown();

    RUN_TESTS();
}
