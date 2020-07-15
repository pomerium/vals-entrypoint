FROM golang:latest as build

# cache depedency downloads
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build

FROM gcr.io/distroless/base
COPY --from=build /vals-entrypoint/vals-entrypoint /bin/vals-entrypoint

ENTRYPOINT [ "/bin/vals-entrypoint" ]