# fuzermount

This is a wrapper tool for `/usr/bin/fusermount3` to control who is calling fusermount with what options.
It's meant to take fusermounts place with SUID and root ownership and if the original call is allowed
it calls the actual fusermount binary to do the work.
This way it's possible to remove SUID/GUID from fusermount completly.
