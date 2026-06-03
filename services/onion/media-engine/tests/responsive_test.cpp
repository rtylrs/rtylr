#include "transform/responsive.hpp"
#include "transform/presets.hpp"
#include "check.hpp"

using namespace voxire::transform;

int main() {
    auto w = breakpointWidths(200, 1000, 5);
    CHECK_EQ(w.size(), static_cast<size_t>(5));
    CHECK_EQ(w.front(), 200);
    CHECK_EQ(w.back(), 1000);
    CHECK_EQ(w[2], 600); // midpoint

    CHECK_EQ(breakpointWidths(100, 800, 1).size(), static_cast<size_t>(1));
    CHECK_EQ(breakpointWidths(100, 800, 1).front(), 800);
    CHECK(breakpointWidths(100, 50, 4).empty());  // inverted range
    CHECK(breakpointWidths(100, 800, 0).empty());

    CHECK(!presetChain("thumb").empty());
    CHECK(!presetChain("avatar").empty());
    CHECK(presetChain("nope").empty());

    RUN_TESTS();
}
