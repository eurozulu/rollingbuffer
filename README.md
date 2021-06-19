# rollingbuffer
a byte buffer which uses a single under lying array, avoiding additional memory allocations / garbage collection.
  
The buffer writes new bytes to the end of the buffer, until it is full, then wraps around and begins filling from the beginning,
filling in the back space freed up from any read operations.  
When no free space is available, writes will return with a zero length until such time as some bytes are read to free space.

