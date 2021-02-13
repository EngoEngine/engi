# How to properly compile and use these

**Note**: This is really for development of the shaders, so unless you're doing that
you can safely ignore this!

### Prerequesites

To do this, you'll need a few tools.

1. Go installed, of course. Latest stuff is test on go1.13.6 darwin/amd64

2. A C-Compiler. For linux and mac, that should come out of the box. On windows,
you'll have to use MinGW or some other way to access GCC or Clang.

3. Vulkan installed (or Molten on Mac)

4. The glslangvalidator tool to compile glsl to spir-V.

5. go-bindata tool installed

6. just run `go generate` in this folder! Yay!
