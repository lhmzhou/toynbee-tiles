FROM lhmzhou.com/toynbee-tiles

ARG URL
ARG PAAS_BUILD_ID
ENV URL=${URL} \
    PAAS_BUILD_ID=${PAAS_BUILD_ID}
WORKDIR $GOPATH/src/lhmzhou/toynbee-tiles

RUN useradd -u 1000 -U -d $GOPATH/src app  && chown -R 1000:1000 $GOPATH

# do anything requiring ROOT above here
USER 1000

RUN echo ${APP_NAME}

ADD . .

ENTRYPOINT ./start.sh
# CMD ["./start"]