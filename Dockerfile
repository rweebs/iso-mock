# Stage build
FROM registry.kudoserver.com/kudo/golang:1.11 as builder

LABEL name="mocking_biller"
LABEL version="1.0.0"

WORKDIR /go/src/bitbucket.org/kudoindonesia/

# Add the keys and set permissions
RUN apk add --no-cache openssh
ARG SSH_PRIVATE_KEY
RUN mkdir ~/.ssh/ && \
    echo "${SSH_PRIVATE_KEY}" > ~/.ssh/id_rsa && \
    chmod 0600 ~/.ssh/id_rsa && \
    touch /root/.ssh/known_hosts && \
    ssh-keyscan bitbucket.org >> /root/.ssh/known_hosts

# git clone source
# ARG ORIGIN
# RUN git clone git@bitbucket.org:kudoindonesia/mocking-biller.git mocking_biller

# RUN cd /go/src/bitbucket.org/kudoindonesia/mocking_biller && git checkout "${ORIGIN}"

WORKDIR /go/src/bitbucket.org/kudoindonesia/mocking_biller
COPY . .

RUN dep ensure -v && GIT_COMMIT=$(git rev-parse --short HEAD) && go build -o ./mocking_biller -ldflags "-X main.GitCommit=$GIT_COMMIT" .

# Stage Runtime Applications
FROM registry.kudoserver.com/kudo/base-image

# # Download Depedencies
RUN apk update && apk add ca-certificates bash jq curl && rm -rf /var/cache/apk/*

# Setting timezone
ENV TZ=Asia/Jakarta
RUN apk add -U tzdata
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

ENV BUILDDIR=/go/src/bitbucket.org/kudoindonesia/mocking_biller

# Add user kudo
RUN adduser -D kudo kudo

# Setting folder workdir
WORKDIR /opt/mocking_biller
RUN mkdir configurations

# Copy Data App
COPY --from=builder $BUILDDIR/mocking_biller mocking_biller
COPY --from=builder $BUILDDIR/configurations/App.yaml.dist /opt/mocking_biller/configurations/App.yaml
COPY --from=builder $BUILDDIR/cert /opt/mocking_biller/cert
# For development in local, you could copy config file from your machine instead
# COPY ./configurations/App.yaml /opt/mocking_biller/configurations/App.yaml
COPY --from=builder $BUILDDIR/run.sh run.sh


# Setting owner file and dir
RUN chmod +x run.sh
RUN chown -R kudo:kudo .

EXPOSE 8089

USER kudo
