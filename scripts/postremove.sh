#!/bin/bash

unlink /usr/bin/fusermount3
mv /opt/fuzermount/fusermount3 /usr/bin/fusermount3
chmod +s /usr/bin/fusermount3
