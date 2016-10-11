* Use logging.
* Functions which searches for README.md in various places and return an fd.
* Use []byte instead of []string. Most stdlib functions require this data type.
* Log file beautifier JS lib.
* Truncate / delete logs.
* Do not throw away HTTP server error status. It may fail to start if trying to start
  the server on a port <1024 when not running as root.
* Implement SSL.
* Read connection info from process descriptor using gopsutil.
* Implement process tree in Go. See e.g. https://play.golang.org/p/VWR9zSoC-p.
* Template caching.
* Navbar in README.md.
* Show number of lines in parenthesis.
* Container info should show container ID, image name and container name. This helps if neither image name nor container name have been set manually by the user.
