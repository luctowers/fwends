mod routes;

use tokio::signal::{self, unix::SignalKind};
use warp::{Filter};

#[tokio::main]
async fn main() {

    let base = warp::get()
        .and(warp::path::end())
        .map(|| "Hello from backend!");

    let api = warp::path("api").and(
        base.or(routes::packs::routes())
    );

    let (addr, server) = warp::serve(api)
        .bind_with_graceful_shutdown(([0, 0, 0, 0], 8080), async {
            let mut terminate = signal::unix::signal(SignalKind::terminate()).unwrap();
            tokio::select! {
                _ = signal::ctrl_c() => {},
                _ = terminate.recv() => {},
            };
        });

    tokio::task::spawn(server).await.ok();
    
}
