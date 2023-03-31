FROM ubuntu:20.04
RUN sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list && \
    sed -i 's/security.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list && \
    apt update && \
    apt install -y --no-install-recommends \
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

ENV LD_LIBRARY_PATH /app/lib:$LD_LIBRARY_PATH
ENV DEBIAN_FRONTEND=noninteractive

RUN mkdir -p /var/log/agora


COPY rtm2-so ./lib
COPY rtm2-wrapper.exe .
RUN chmod 755 ./rtm2-wrapper.exe
