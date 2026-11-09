//! Spike entrypoint: build the router and prove it compiles standalone.
//! HTTP behaviour is proven by `cargo test -p spike` (in-process).

use spike::spike_router;

fn main() {
    let _router = spike_router();
    eprintln!("spike router built (2 routes)");
}
