use serde::{Serialize};
use warp::{Filter, filters::BoxedFilter, Reply};

#[derive(Serialize, Clone)]
struct Pack {
    id: u64,
    name: String,
    roleCount: u16,
    stringCount: u16,
}

pub fn routes() -> BoxedFilter<(impl Reply,)> {
  let packsData: Vec<Pack> = vec![
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

  warp::get()
    .and(warp::path("packs"))
    .map(move || warp::reply::json(&packsData))
    .boxed()
}
