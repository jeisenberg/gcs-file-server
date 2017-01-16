FROM google/appengine-go

RUN apt-get update

RUN apt-get install coreutils

RUN apt-get install --no-install-recommends -y -q \
    curl build-essential git mercurial bzr

RUN sha1sum

RUN mkdir /goroot && curl https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz | tar xvzf - -C /goroot --strip-components=1
RUN mkdir /gopath

ENV GOROOT /goroot
ENV GOPATH /gopath
ENV PATH $PATH:$GOROOT/bin:$GOPATH/bin

RUN go get code.google.com/p/goauth2/oauth

RUN go get google.golang.org/appengine

RUN go get -u	golang.org/x/net/context
RUN go get -u	golang.org/x/oauth2
RUN go get -u	golang.org/x/oauth2/google
RUN go get -u 	google.golang.org/cloud
RUN go get -u   google.golang.org/cloud/storage

ADD . /app
RUN /bin/bash /app/_ah/build.sh
