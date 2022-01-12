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
ARG LOG_DEBUG=true
ENV LOG_DEBUG=${LOG_DEBUG}
ARG HTTP_DEBUG=true
ENV HTTP_DEBUG=${HTTP_DEBUG}
CMD ["fwends-backend"]
