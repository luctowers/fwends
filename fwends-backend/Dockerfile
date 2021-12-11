FROM rust:1.57-alpine3.14 as builder
RUN apk add musl-dev
WORKDIR /usr/src/
RUN cargo new fwends-backend
WORKDIR /usr/src/fwends-backend
COPY Cargo.toml Cargo.lock ./
RUN cargo build --release
COPY src src
# touching main.rs is needed because of https://github.com/rust-lang/cargo/issues/6529
RUN touch src/main.rs
RUN cargo install --path .

FROM alpine:3.14
EXPOSE 8080
COPY --from=builder /usr/local/cargo/bin/fwends-backend /usr/local/bin/fwends-backend
CMD ["fwends-backend"]
