FROM golang:1.17-alpine AS build
WORKDIR /usr/src/fwends-backend
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN go build -gcflags="${SKAFFOLD_GO_GCFLAGS}" -o ./fwends-backend

FROM alpine
COPY --from=build /usr/src/fwends-backend/fwends-backend /usr/local/bin/fwends-backend
EXPOSE 80
CMD ["fwends-backend"]
