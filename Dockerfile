FROM    golang:1.8
#RUN    go get github.com/Masterminds/glide
RUN     go get -u -v github.com/xkeyideal/glide
RUN     apt-get update
RUN     apt-get install -y --no-install-recommends apt-utils && apt-get install -y git
WORKDIR $GOPATH/src/github.com/paddlepaddle
RUN     mkdir -p $GOPATH/src/github.com/paddlepaddle/paddlejob
# Add ENV http_proxy=[your proxy server] if needed
# run glide install before building go sources, so that
# if we change the code and rebuild the image can cost little time
ADD     ./glide.yaml ./glide.lock $GOPATH/src/github.com/paddlepaddle/paddlejob/
WORKDIR $GOPATH/src/github.com/paddlepaddle/paddlejob
RUN mkdir -p ~/.glide && touch ~/.glide/mirrors.yaml && glide mirror set https://golang.org/x/crypto https://github.com/golang/crypto --vcs git
RUN glide mirror set https://golang.org/x/net https://github.com/golang/net --vcs git
RUN glide mirror set https://golang.org/x/text https://github.com/golang/text --vcs git
RUN glide mirror set https://golang.org/x/sys https://github.com/golang/sys --vcs git
RUN     glide install --strip-vendor
ADD     . $GOPATH/src/github.com/paddlepaddle/paddlejob
RUN     go build -o /usr/local/bin/paddlejob github.com/paddlepaddle/paddlejob/cmd/paddlejob
RUN     rm -rf $GOPATH/src/github.com/paddlepaddle/paddlejob
ENTRYPOINT ["paddlejob", "--alsologtostderr"]
