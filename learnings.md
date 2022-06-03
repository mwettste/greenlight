# CORS - Cross Origin Requests
If two URLs have the same scheme, host and port they are said to be from the same origin.

CORS does NOT block the following:
* a webpage can embed certain resources from another origin in its HTML
* a webpage can *send* data to a different origin

CORS DOES block the following:
* a webpage on one origin is not allowed to receive/read data from another origin

*Important*: sending of cross-origin data is allowed, which is why CSRF is possible and we need to take additional measures to prevent them.

CORS request are classified as simple, when they meet all the following criteria:
* HTTP request method is one of the CORS-safe methods: HEAD / GET / POST
* a CORS safe listed header is used: Accept, Accept-Language, Content-Language, Content-Type
* the value of Content-Type header (if set) is one of the following: application/x-www-form-urlencoded, multipart/form-data, text/plain

When a request is not classified as simple, an initial preflight requst is made.
A preflight request is made by setting certain headers to test if the actual request would be allowed or not:
* Access-Control-Request-Method
* Access-Control-Request-Headers
* Origin