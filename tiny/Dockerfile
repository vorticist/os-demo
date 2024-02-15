FROM ubuntu:latest AS base
WORKDIR /ui

RUN  apt update && apt install -y git curl build-essential python3-dev pip bash

RUN git clone https://github.com/oobabooga/text-generation-webui.git

WORKDIR /ui/text-generation-webui
SHELL ["/bin/bash", "-c"]

RUN sed -i "s/server_name=None if not shared.args.listen else (shared.args.listen_host or '0.0.0.0'),/server_name='0.0.0.0',/g" "server.py"

RUN chmod a+x start_linux.sh
RUN MODEL=TinyLlama_TinyLlama-1.1B-step-50K-105b GPU_CHOICE=A USE_CUDA118=FALSE LAUNCH_AFTER_INSTALL=FALSE INSTALL_EXTENSIONS=FALSE ./start_linux.sh
RUN pip install -r requirements.txt
RUN python3 download-model.py TinyLlama/TinyLlama-1.1B-step-50K-105b
EXPOSE ${CONTAINER_PORT:-7860} ${CONTAINER_API_PORT:-5000} ${CONTAINER_API_STREAM_PORT:-5005}

CMD umask 0002 && export HOME=/ui/text-generation-webui && ./start_linux.sh