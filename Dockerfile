FROM golang:latest as build

WORKDIR /vals-entrypoint
# cache depedency downloads
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build

FROM gcr.io/distroless/static
COPY --from=build /vals-entrypoint/vals-entrypoint /bin/vals-entrypoint

ENTRYPOINT [ "/bin/vals-entrypoint" ]