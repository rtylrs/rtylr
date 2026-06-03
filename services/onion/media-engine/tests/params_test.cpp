#include "transform/params.hpp"
#include "check.hpp"

using namespace voxire::transform;

int main() {
    auto steps = splitSteps("w_300,c_fill/e_blur:200/f_auto");
    CHECK_EQ(steps.size(), static_cast<size_t>(3));
    CHECK_EQ(steps[0], std::string("w_300,c_fill"));
    CHECK_EQ(steps[1], std::string("e_blur:200"));

    // Leading/trailing slashes and empty steps are dropped.
    CHECK_EQ(splitSteps("/a//b/").size(), static_cast<size_t>(2));

    auto p = parseParams("w_300,h_200,c_fill,g_auto,e_blur:200");
    CHECK_EQ(p["w"], std::string("300"));
    CHECK_EQ(p["h"], std::string("200"));
    CHECK_EQ(p["c"], std::string("fill"));
    CHECK_EQ(p["g"], std::string("auto"));
    // Effect value keeps everything after the first underscore, colon included.
    CHECK_EQ(p["e"], std::string("blur:200"));

    auto bare = parseParams("a_1_2");
    CHECK_EQ(bare["a"], std::string("1_2")); // only the first underscore splits

    RUN_TESTS();
}
