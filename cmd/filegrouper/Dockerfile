FROM centos:7
MAINTAINER Stephen Ingram stevei@spaceflightinductries.com

ADD filegrouper .
RUN mkdir /opt/grouper && mv filegrouper /opt/grouper/filegrouper && chmod +x /opt/grouper/filegrouper

WORKDIR /opt/grouper
ENTRYPOINT [ "/opt/grouper/filegrouper" ]
