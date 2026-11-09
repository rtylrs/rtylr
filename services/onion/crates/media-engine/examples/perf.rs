//! Perf harness: median/min over the primitive parses, mirroring
//! `media-engine/bench/benchmark.cpp` (run before/after a change).
//! `cargo run -p media-engine --example perf`

use std::time::Instant;

fn median_us(iters: usize, mut work: impl FnMut()) -> (f64, f64) {
    let mut samples = Vec::with_capacity(iters);
    for _ in 0..iters {
        let at = Instant::now();
        work();
        samples.push(at.elapsed().as_secs_f64() * 1e6);
    }
    samples.sort_by(|a, b| a.partial_cmp(b).unwrap());
    let median = samples[samples.len() / 2];
    let min = samples.first().copied().unwrap_or(0.0);
    (median, min)
}

fn main() {
    let iters: usize = std::env::args()
        .nth(1)
        .and_then(|text| text.parse().ok())
        .unwrap_or(7);
    let (median, min) = median_us(iters, || {
        let _ = media_engine::parse_size("256x256");
        let _ = media_engine::parse_color("#ff8800");
        let _ = media_engine::offset_for(media_engine::Gravity::Center, 100, 100, 300, 300);
    });
    println!("primitives iters={iters} median={median:.1}us min={min:.1}us");
}
