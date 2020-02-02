FROM golang:1.13

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o pp ./cmd

FROM scratch

WORKDIR /
COPY --from=0 /app/pp .

ENTRYPOINT ["/pp"]
