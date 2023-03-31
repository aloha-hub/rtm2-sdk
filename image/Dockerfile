FROM ubuntu:20.04
RUN sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list
RUN sed -i 's/security.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list
RUN apt update
RUN apt install -y --no-install-recommends \
    curl \
    gdb \
    git \
    g++ \
    g++-multilib \
    ca-certificates \
    libtool \
    locales \
    make \
    pkg-config \
    libxdamage1 \
    libxcomposite-dev \
    libdrm-dev \
    apt-file \
    libc++-dev \
    libc++1

WORKDIR /app

RUN mkdir -p /var/log/agora

# copy so library
COPY rtm2-so/libagora-fdkaac.so ./lib/
COPY rtm2-so/libagora-ffmpeg.so ./lib/
COPY rtm2-so/libagora-soundtouch.so ./lib/
# set path
ENV LD_LIBRARY_PATH /app/lib:$LD_LIBRARY_PATH
ENV DEBIAN_FRONTEND=noninteractive

COPY rtm2-so/libagora_rtc_sdk.so ./lib/

COPY rtm2-wrapper.exe .
RUN chmod 755 ./rtm2-wrapper.exe
