use serde::Serialize;
use warp::{filters::BoxedFilter, Filter, Reply};

#[derive(Serialize, Clone)]
struct Pack {
    id: u64,
    name: String,
    #[serde(rename = "roleCount")]
    role_count: u16,
    #[serde(rename = "stringCount")]
    string_count: u16,
}

pub fn routes() -> BoxedFilter<(impl Reply,)> {
    let pack_data: Vec<Pack> = vec![
        Pack {
            id: 0,
            name: String::from("Bar Pack One"),
            role_count: 3,
            string_count: 27,
        },
        Pack {
            id: 1,
            name: String::from("Foo Pack Two"),
            role_count: 4,
            string_count: 32,
        },
    ];

    warp::get()
        .and(warp::path("packs"))
        .map(move || warp::reply::json(&pack_data))
        .boxed()
}
