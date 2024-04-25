FROM golang:1.22.2 AS build-stage
WORKDIR /app
COPY go.mod go.sum /
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /acid-ssh

FROM gcr.io/distroless/base-debian11 AS build-release-stage
WORKDIR /
COPY --from=build-stage /acid-ssh /acid-ssh
ENTRYPOINT ["/acid-ssh"]
