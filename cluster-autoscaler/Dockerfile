FROM golang:1.17.5 AS build

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
RUN go build -o cluster-autoscaler

FROM gcr.io/distroless/static:latest

COPY --from=build /gopath/src/k8s.io/autoscaler/cluster-autoscaler/cluster-autoscaler /
CMD ["/cluster-autoscaler"]
