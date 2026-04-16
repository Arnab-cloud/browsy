# Browsy : A tiny browser in Go

### TODO:

1. Create a Request struct
2. Create a Reponse struct
3. Sperate request from response
4. 1-5 view-source. Add support for the view-source scheme; navigating to view-source:http://example.org/ should show the HTML source instead of the rendered page. Add support for this scheme. Your browser should print the entire HTML file as if it was text. You’ll want to have also implemented Exercise 1-4.
5. 1-6 Keep-alive. Implement Exercise 1-1; however, do not send the Connection: close header (send Connection: keep-alive instead). When reading the body from the socket, only read as many bytes as given in the Content-Length header and don’t close the socket afterward. Instead, save the socket, and if another request is made to the same server reuse the same socket instead of creating a new one. (You’ll also need to pass the "rb" option to makefile or the value reported by Content-Length might not match the length of the string you’re reading.) This will speed up repeated requests to the same server, which are common.
