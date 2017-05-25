FROM centos:7
MAINTAINER Stephen Ingram stevei@spaceflightinductries.com

ADD extractbucket .
RUN mkdir /opt/bucket && mv extractbucket /opt/bucket/extractbucket && chmod +x /opt/bucket/extractbucket

WORKDIR /opt/bucket
ENTRYPOINT [ "/opt/bucket/extractbucket" ]
