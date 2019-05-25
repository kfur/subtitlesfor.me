FROM ubuntu:18.04

EXPOSE 80

WORKDIR /root/go/src/github.com/kfur/subtitler
COPY . .

RUN apt update && apt upgrade -y && apt install -y software-properties-common && \
apt update && add-apt-repository -s ppa:jonathonf/ffmpeg-4 && \
apt install -y golang ffmpeg libavformat-dev libavutil-dev libavfilter-dev \
libavdevice-dev libswscale-dev libswresample-dev && \
go build -ldflags "-s" && \
apt remove -y golang && \
apt-get autoremove -y && apt-get clean && apt-get autoclean

CMD ["./subtitler"]