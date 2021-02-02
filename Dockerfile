FROM golang:1.15.7-buster
LABEL maintainer="Rajdeep Bharati <rajdeep@buidllabs.io>"

RUN apt-get update
# RUN apt-cache depends ffmpeg
# RUN apt-get install -y ffmpeg
# RUN ffmpeg -version
# RUN apt-get install -y sqlite3

RUN mkdir /filecoin-chain-indexer
WORKDIR /filecoin-chain-indexer

# COPY go.mod go.sum ./
COPY . .
RUN git submodule update --init --recursive
RUN go mod download
RUN go build main.go

ENV FULLNODE_API_INFO=${FULLNODE_API_INFO}
ENV LOTUS_RPC_ENDPOINT=${LOTUS_RPC_ENDPOINT}
ENV DB=${DB}
ENV FROM=${FROM}
ENV TO=${TO}
ENV PORT=${PORT}
ENV POW_TOKEN=${POW_TOKEN}
ENV POWERGATE_ADDR=${POWERGATE_ADDR}

EXPOSE ${PORT}

# ENTRYPOINT [ "./main", "--cmd=migrate" ]

CMD [ "./main", "--cmd=index" ]
