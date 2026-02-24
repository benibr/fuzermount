#!/bin/bash

mkdir -p /opt/fuzermount/

mv /usr/bin/fusermount3 /opt/fuzermount/fusermount3
chmod -s /opt/fuzermount/fusermount3

ln -sf /opt/fuzermount/fuzermount /usr/bin/fusermount3
