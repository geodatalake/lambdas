FROM centos:7
MAINTAINER Steve Ingram stevei@spaceflightinductries.com

ADD processgeo .
RUN mkdir /opt/processgeo && mv processgeo /opt/processgeo/processgeo && chmod +x /opt/processgeo/processgeo

ADD EPSGTITLE.bin .
ADD EPSGVAL.bin .
RUN mkdir /opt/reproject && mkdir /opt/reproject/bins && mv EPSG*.bin /opt/reproject/bins

WORKDIR /opt/processgeo
ENTRYPOINT [ "/opt/processgeo/processgeo" ]
