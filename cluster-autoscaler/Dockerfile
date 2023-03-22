FROM golang:1.19.13 AS build

ARG TARGETARCH
ENV GOPATH /gopath/
ENV PATH $GOPATH/bin:$PATH
ENV GO111MODULE auto
ENV GOARCH ${TARGETARCH}

RUN apt-get update && apt-get --yes install libseccomp-dev
RUN go version
RUN go get github.com/tools/godep
RUN godep version

WORKDIR /gopath/src/k8s.io/autoscaler/cluster-autoscaler
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -o cluster-autoscaler --ldflags "-s"

FROM alpine
COPY --from=build /gopath/src/k8s.io/autoscaler/cluster-autoscaler/cluster-autoscaler /

CMD ["./cluster-autoscaler"]