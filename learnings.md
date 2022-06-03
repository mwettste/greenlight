# CORS - Cross Origin Requests
If two URLs have the same scheme, host and port they are said to be from the same origin.

CORS does NOT block the following:
* a webpage can embed certain resources from another origin in its HTML
* a webpage can *send* data to a different origin

CORS DOES block the following:
* a webpage on one origin is not allowed to receive/read data from another origin

*Important*: sending of cross-origin data is allowed, which is why CSRF is possible and we need to take additional measures to prevent them.
