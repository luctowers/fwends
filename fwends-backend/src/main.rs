use tokio::signal::{self, unix::SignalKind};
use warp::{Filter};
use serde::{Serialize};

#[derive(Serialize, Clone)]
struct Pack {
    id: u64,
    name: String,
    roleCount: u16,
    stringCount: u16,
}

#[tokio::main]
async fn main() {

    let base = warp::get()
        .and(warp::path!("api"))
        .map(|| "Hello from backend!");

    let packs = vec![
        Pack {
            id: 0,
            name: String::from("Bar Pack One"),
            roleCount: 3,
            stringCount: 27
        },
        Pack {
            id: 1,
            name: String::from("Foo Pack Two"),
            roleCount: 4,
            stringCount: 32
        },
    ];

    let packs = warp::get()
        .and(warp::path!("api" / "packs"))
        .map(move || warp::reply::json(&packs));

    let api = base.or(packs);

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
