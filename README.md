# goLru

A LRU cache with expiration checking for golang.

## API

* NewLRUCache

  return a new LRU cache. Input a integer parameter as the capacity of cache.

* Tail

  return the tail of cache.

* Count

  return the count of items in cache.

* OperateForAll

  input a function clousure to do it for all item in the cache.

* Delete

  delete a key/value pair. Input a key as parameter.

* Exists

  judge inputted key whether it is exist.

* Put

  put a key/value pair into cache. If this key exists, update the value.

* Get

  get the value of inputted key. Return  ``nil`` if the key doesn't exist.

* Flush

  delete all key/value pair in the cache.
