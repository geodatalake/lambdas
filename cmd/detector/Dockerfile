FROM centos:7
MAINTAINER Stephen Ingram stevei@spaceflightinductries.com

ADD detector
RUN mkdir /opt/detect && mv detector /opt/detect/detector && chmod +x /opt/detect/detector

WORKDIR /opt/detect
ENTRYPOINT [ "/opt/detect/detector" ]
