#FROM golang:alpine as builder
#ARG GOROOT
#ENV GOROOT $GOROOT
#RUN mkdir ./build
#WORKDIR ./build
#COPY ./ ./build
#RUN CGO_ENABLED=0 GOOS=linux \
#    go build -a -installsuffix cgo -v -o ./bin/keystonegrpcadapter ./build/cmd/main.go
#
#FROM alpine:3.8
#RUN apk --no-cache add ca-certificates
#WORKDIR /bin/
#COPY --from=builder ./bin/keystonegrpcadapter .
#ENTRYPOINT [ "/bin/keystonegrpcadapter" ]
#CMD [ "9090" ]
#EXPOSE 9090


FROM bradobro/golang-alpine-dep as builder


WORKDIR /go/src/istio.io/istio/mixer/adapter/keystonegrpcadapter
COPY . ./
#RUN apk add --update --no-cache make bash curl git ca-certificates openssh-client
#RUN curl -fsSL -o /usr/local/bin/dep  `curl -fsSL https://api.github.com/repos/golang/dep/releases/latest`
#RUN chmod +x /usr/local/bin/dep
RUN dep ensure -v
RUN go build -o build/keystonegrpcadapter ./cmd/

# Copy the binary
FROM alpine

RUN adduser -S -D -H -h /app appuser

USER appuser
COPY --from=builder //go/src/istio.io/istio/mixer/adapter/keystonegrpcadapter/build/keystonegrpcadapter /app/
WORKDIR /app
ENTRYPOINT ["./keystonegrpcadapter"]
CMD [ "9090" ]
EXPOSE 9090
