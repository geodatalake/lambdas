FROM centos:7
MAINTAINER Chris Mangold chrism@spaceflightinductries.com

ADD detector .
RUN mkdir /opt/reproject && mv reprojectbounds /opt/reproject/reprojectbounds && chmod +x /opt/reproject/reprojectbounds

WORKDIR /opt/reproject
ENTRYPOINT [ "/opt/reproject/reprojectbounds" ]