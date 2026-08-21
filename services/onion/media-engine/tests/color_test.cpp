#include "color.hpp"
#include "check.hpp"

using namespace voxire;

int main() {
    Color hex6 = parseColor("#ff8800");
    CHECK_EQ(hex6.r, 255);
    CHECK_EQ(hex6.g, 136);
    CHECK_EQ(hex6.b, 0);
    CHECK_EQ(hex6.a, 255);

    Color noHash = parseColor("ff8800");
    CHECK_EQ(noHash.r, 255);

    Color hex8 = parseColor("#11223380");
    CHECK_EQ(hex8.r, 0x11);
    CHECK_EQ(hex8.a, 0x80);

    Color shortRGB = parseColor("#f80");
    CHECK_EQ(shortRGB.r, 255);
    CHECK_EQ(shortRGB.g, 136);
    CHECK_EQ(shortRGB.b, 0);

    CHECK_EQ(parseColor("black").r, 0);
    CHECK_EQ(parseColor("white").r, 255);
    CHECK_EQ(parseColor("transparent").a, 0);
    CHECK_EQ(parseColor("red").r, 255);
    CHECK_EQ(parseColor("red").g, 0);
    CHECK_EQ(parseColor("navy").b, 128);
    CHECK_EQ(parseColor("grey").r, parseColor("gray").r);

    // bands() drops alpha when opaque, keeps it when translucent.
    CHECK_EQ(parseColor("white").bands().size(), static_cast<size_t>(3));
    CHECK_EQ(parseColor("#00000080").bands().size(), static_cast<size_t>(4));

    CHECK_THROWS(parseColor("#ggg"));
    CHECK_THROWS(parseColor("nope"));

    RUN_TESTS();
}
