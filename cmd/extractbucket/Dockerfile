FROM centos:7
MAINTAINER Stephen Ingram stevei@spaceflightinductries.com

ADD extractbucket .
RUN mkdir /opt/bucket && mv extractbucket /opt/bucket/extractbucket && chmod +x /opt/bucket/extractbucket
RUN mkdir /root/.aws

RUN ls -al /root/.aws
RUN cd && pwd

WORKDIR /opt/bucket
ENTRYPOINT [ "/opt/bucket/extractbucket" ]
