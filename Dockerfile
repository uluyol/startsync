FROM scratch
MAINTAINER Muhammed Uluyol <uluyol0@gmail.com>

ADD startsyncd /
ADD startsync /

ENTRYPOINT ["/startsyncd"]
