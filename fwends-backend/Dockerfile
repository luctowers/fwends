FROM rust:1.56-alpine3.14 as builder
RUN apk add musl-dev
RUN cargo install cargo-watch
WORKDIR /usr/src/
RUN cargo new fwends-backend
WORKDIR /usr/src/fwends-backend
COPY Cargo.toml Cargo.lock ./
RUN cargo build
COPY src src
RUN touch src/main.rs
RUN cargo build
CMD [ "cargo", "watch", "-x",  "run" ]
