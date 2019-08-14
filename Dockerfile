FROM golang:1.12-buster as builder

# Copy in the go src
WORKDIR /go/src/github.com/aflc/extended-resource-toleration-webhook

ENV GO111MODULE="on"
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ert-webhook .

# Copy the controller-manager into a thin image
FROM alpine:3.10
WORKDIR /
COPY --from=builder /go/src/github.com/aflc/extended-resource-toleration-webhook/ert-webhook .
ENTRYPOINT ["/ert-webhook"]
