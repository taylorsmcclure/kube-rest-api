FROM golang:1.18
#Build args
ENV PORT=8443
ENV REDIS_ADDR="redis-master.redis.svc.cluster.local:6379"

# kube-server user
ARG USERNAME=kube-server
ARG USER_UID=1000
ARG USER_GID=$USER_UID

#Adding Labels of build
LABEL maintainer="Taylor McClure"
LABEL repo="https://github.com/taylorsmcclure/kube-server"

RUN groupadd --gid $USER_GID $USERNAME && \
    useradd --uid $USER_UID --gid $USER_GID -m $USERNAME && \
    mkdir /kube-server && \
    chmod -R 775 /kube-server && \
    chown -R $USERNAME:$USERNAME /kube-server
COPY cmd/kube-server/bin/ /kube-server
COPY scripts/entrypoint.sh /kube-server
COPY scripts/liveness.sh /kube-server
WORKDIR /kube-server

EXPOSE $PORT

USER $USERNAME

#Execution
ENTRYPOINT [ "./entrypoint.sh" ]
