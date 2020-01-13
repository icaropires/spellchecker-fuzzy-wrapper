FROM golang:1.13-alpine3.11 as builder

WORKDIR /spellchecker

COPY service_http.go .

RUN apk --no-cache add git \
    && go get -v "github.com/sajari/fuzzy" \
    && go build service_http.go


FROM alpine:3.11

WORKDIR /spellchecker

COPY --from=builder /spellchecker/service_http .

ENTRYPOINT ["/spellchecker/service_http"]
