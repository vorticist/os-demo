FROM golang:alpine3.19 as builder
COPY . /server
WORKDIR /server

RUN go build -o server .
RUN ls

FROM ultralytics/ultralytics:latest as yolov8
WORKDIR /server
COPY --from=builder /server/server .
COPY --from=builder /server/client ./client
COPY --from=builder /server/static ./static

RUN ls


ENV TZ=US/Pacific
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN apt update && apt upgrade -y
RUN conda update -y ffmpeg

ENTRYPOINT [ "/server/server" ]