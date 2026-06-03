cmake_minimum_required(VERSION 3.16)
project(media-engine VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Release build by default
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

set(CMAKE_CXX_FLAGS_RELEASE "-O2 -DNDEBUG")

# ── Dependencies ─────────────────────────────────────────────────────────────

find_package(PkgConfig REQUIRED)

# libvips C++ bindings
pkg_check_modules(VIPS REQUIRED vips-cpp)

# FFmpeg (optional — used by video_processor when activated)
pkg_check_modules(FFMPEG QUIET
    libavcodec
    libavformat
    libavutil
    libswscale
)

# ── Sources ───────────────────────────────────────────────────────────────────

# Glob every translation unit under src/ (top level plus ops/ and video/) so new
# operations are picked up automatically. CONFIGURE_DEPENDS re-globs on build.
file(GLOB_RECURSE SOURCES CONFIGURE_DEPENDS "${CMAKE_CURRENT_SOURCE_DIR}/src/*.cpp")

add_executable(media-engine ${SOURCES})

target_include_directories(media-engine PRIVATE
    include
    ${VIPS_INCLUDE_DIRS}
)

target_link_libraries(media-engine PRIVATE
    ${VIPS_LIBRARIES}
)

target_compile_options(media-engine PRIVATE
    ${VIPS_CFLAGS_OTHER}
    -Wall
    -Wextra
    -Wpedantic
)

# Surface the project version to the binary (used by the `version` subcommand).
target_compile_definitions(media-engine PRIVATE
    MEDIA_ENGINE_VERSION="${PROJECT_VERSION}"
)

if(FFMPEG_FOUND)
    target_compile_definitions(media-engine PRIVATE VOXIRE_FFMPEG_AVAILABLE)
    target_include_directories(media-engine PRIVATE ${FFMPEG_INCLUDE_DIRS})
    target_link_libraries(media-engine PRIVATE ${FFMPEG_LIBRARIES})
    message(STATUS "FFmpeg found — video processing support enabled")
else()
    message(STATUS "FFmpeg not found — video processing will be scaffolded only")
endif()

# ── Install ───────────────────────────────────────────────────────────────────

install(TARGETS media-engine DESTINATION bin)
