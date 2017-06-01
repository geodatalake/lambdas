FROM centos:7
MAINTAINER Chris Mangold chrism@spaceflightinductries.com

ADD reprojectbounds .
RUN mkdir /opt/reproject && mv reprojectbounds /opt/reproject/reprojectbounds && chmod +x /opt/reproject/reprojectbounds

ADD EPSGTITLE.bin .
ADD EPSGVAL.bin .
RUN mkdir /opt/reproject/bins && mv EPSG*.bin /opt/reproject/bins

WORKDIR /opt/reproject
ENTRYPOINT [ "/opt/reproject/reprojectbounds" ]
