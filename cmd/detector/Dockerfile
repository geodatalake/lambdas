FROM centos:7
MAINTAINER Stephen Ingram stevei@spaceflightinductries.com

ADD detector .
RUN mkdir /opt/detect && mv detector /opt/detect/detector && chmod +x /opt/detect/detector

ADD EPSGTITLE.bin .
ADD EPSGVAL.bin .
RUN mkdir /opt/reproject/bins && mv EPSG*.bin /opt/reproject/bins

WORKDIR /opt/detect
ENTRYPOINT [ "/opt/detect/detector" ]
