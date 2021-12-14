FROM golang:1.17-alpine AS build
WORKDIR /usr/src/app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /fwends-backend

FROM alpine
COPY --from=build /usr/src/app/fwends-backend /usr/local/bin/fwends-backend
EXPOSE 8080
CMD ["./fwends-backend"]
