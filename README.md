# servicewrapper

This is a small helper tool that wraps a Windows service for debugging purposes.
It can be used to:
* Add a delay before the wrapped service is launched, allowing to attach a
debugger, for example [x64dbg](https://github.com/x64dbg/x64dbg) using the
[DbgChild](https://github.com/David-Reguera-Garcia-Dreg/DbgChild) plugin.
* Redirect the service's stdout and stderr to a file. Useful to capture errors
(useful to catch `panic` terminations in Golang services).

### Flags

`-delay <milliseconds>`: Wait this number of milliseconds before launching service 

`-output <path>`: Redirect stdout/stderr to the given file. Relative paths
will be interpreted relative to the SYSTEM32 folder.

### Usage

Register your service as:

> /path/to/servicewrapper.exe [flags] /path/to/original/service.exe optional arguments
