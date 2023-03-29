FROM archlinux:base

RUN echo "Server = https://mirrors.ustc.edu.cn/archlinux/\$repo/os/\$arch" > /etc/pacman.d/mirrorlist
RUN echo "Server = https://mirrors.tuna.tsinghua.edu.cn/archlinux/\$repo/os/\$arch" >> /etc/pacman.d/mirrorlist
RUN pacman-key --init
RUN pacman --noconfirm -Sy archlinux-keyring
RUN pacman --noconfirm -Syu make git gcc go llvm14

WORKDIR /sim
COPY . .
RUN export GOPROXY="https://goproxy.cn,direct" && go mod download
RUN make build