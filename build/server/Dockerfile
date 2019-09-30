FROM cargo.caicloud.xyz/library/debian:stretch

# Set the default timezone to Shanghai
RUN echo "Asia/Shanghai" > /etc/timezone
RUN dpkg-reconfigure -f noninteractive tzdata

WORKDIR /caicloud/compass/admission

ADD bin/server /caicloud/compass/admission/

ADD etc/review.yaml /caicloud/compass/admission/

CMD ["/caicloud/compass/admission/server"]
