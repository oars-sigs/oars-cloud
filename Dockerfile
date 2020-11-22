FROM library/centos:7
ADD bin/oars /bin/
ENTRYPOINT ["/bin/oars"]
