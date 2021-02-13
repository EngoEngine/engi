# Engo
[![GoDoc](https://godoc.org/github.com/EngoEngine/engo?status.svg)](https://godoc.org/github.com/EngoEngine/engo)
[![Join the chat at https://gitter.im/EngoEngine/engo](https://badges.gitter.im/EngoEngine/engo.svg)](https://gitter.im/EngoEngine/engo?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) ![License](https://img.shields.io/badge/License-MIT-blue.svg)
[![Build Status](https://travis-ci.org/EngoEngine/engo.svg?branch=master)](https://travis-ci.org/EngoEngine/engo)
[![Build status](https://ci.appveyor.com/api/projects/status/019qc8hncmhnje83?svg=true)](https://ci.appveyor.com/project/otraore/engo)
[![Go Report Card](https://goreportcard.com/badge/github.com/EngoEngine/engo)](https://goreportcard.com/report/github.com/EngoEngine/engo)
[![Coverage Status](https://coveralls.io/repos/github/EngoEngine/engo/badge.svg?branch=master)](https://coveralls.io/github/EngoEngine/engo?branch=master)

A cross-platform game engine written in Go following an interpretation of the Entity Component System paradigm. Engo is
currently compilable for Mac OSX, Linux and Windows. With the release of Go 1.4, supporting Android and the inception of
iOS compatibility, mobile has been be added as a release target. Web support
([wasm](https://github.com/golang/go/wiki/WebAssembly)) is also available.

v1.0 is now available! To celebrate, there will be a game jam coming soon to celebrate the release, start actually
building things and hopefully find any issues. Updates for this will come soon.

## Getting in touch / Contributing

We have a [gitter](https://gitter.im/EngoEngine/engo) chat for people to join who want to further discuss `engo`. We are happy to discuss bugs, feature requests and would love to hear about the projects you are building!

## Getting Started

### Theory: `common` vs `engo`

There are currently two major important packages within this repository: `github.com/EngoEngine/engo` and `github.com/EngoEngine/engo/common`.

The top level `engo` package contains the functionality of creating windows, starting the game, creating an OpenGL
context and handling input. It is designed to be used with Systems designed as per `github.com/EngoEngine/ecs` specifications.
The `common` package contains our ECS implementations of common game development Systems like a  `RenderSystem` or
`CameraSystem`.

### Practice: Getting it to Run

1. First, you have to install some dependencies:
  1. If you're running on Debian/Ubuntu:
    `sudo apt-get install libasound2-dev libglu1-mesa-dev freeglut3-dev mesa-common-dev xorg-dev libgl1-mesa-dev git-all`
  2. If you're running on Windows you'll need a gcc compiler that the go tool can use and have `gcc.exe` in your PATH environmental variable. We recommend [Mingw](http://mingw-w64.org/doku.php/start) since it has been tested. You'll also need git installed, we recommend getting it from [The official Git site](http://git-scm.com/download/win)
  3. If you're on OSX, you will also need Git. You can find instructions [here](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git#Installing-on-Mac). You can also use homebrew to install git as well. [Open an issue if you have any issues](https://github.com/EngoEngine/engo/issues/new)
2. Then, you can go get it:
`go get -u github.com/EngoEngine/engo`
  1. You may also want to get the dependencies of platform specific builds, so that build tools like godef can use them:
  `go get -u -tags js ./...`
  `go get -u -tags android ./...`
3. Now, you have two choices:
  1. Visit [our website](https://engoengine.github.io/), which hosts a full-blown tutorial series on how to create your own game, and on top of that, has some conceptual explanations;
  2. Check out some demos in our [demos folder](https://github.com/EngoEngine/engo/tree/master/demos).
4. Finally, if you run into problems, if you've encountered a bug, or want to request a feature, feel free to shoot
us a DM or [create an issue](https://github.com/EngoEngine/engo/issues/new).

## Build Tags
There are several different build tags available to swap out the renderer as well as
the Window Systems. Currently, the tags are as follows

| tag | what it does | requirements |
|:---:|---|---|
| -mobilebind | Builds the file as an .aar library to link to an apk. For building mobile apps that require functionality not provided by gomobile, such as 3rd party O-Auth. | The ability to build apk files, either through Android Studio, maven, or some other way. |
| -sdl | Uses SDL rather than glfw to build the window system. | SDL installed. A good walkthrough can be found [here](https://github.com/veandco/go-sdl2#requirements) |
| -vulkan | Uses Vulkan as the renderer rather than OpenGL. Does not work on mobile (yet!). | Vulkan installed. How to install on [MacOS](https://github.com/KhronosGroup/MoltenVK#building-from-the-command-line), [Windows](https://vulkan.lunarg.com/doc/view/1.1.121.2/windows/getting_started.html), and [Linux](https://vulkan.lunarg.com/doc/view/1.1.114.0/linux/getting_started.html) |
| -headless | Does not build the OpenGL context at all. For use on servers or other places where a framebuffer isn't accessible. Use in CONJUNCTION with the Headless Run Option! | No requirements |

## Breaking Changes Since v1.0
Engo is always undergoing a lot of optimizations and constantly gets new features. However, this sometimes means things break. In order to make transitioning easier for you,
we have a list of those changes, with the most recent being at the top. If you run into any problems, please contact us at [gitter](https://gitter.im/EngoEngine/engo).

* common.RenderSystem's BufferContent and Buffer that allow access to the render buffer for custom shaders has been
wrapped into BufferData. This allows access to OpenGL and Vulkan specific buffers for custom shaders.
* common.Drawable has been changed so that it does not contain any references to the specific renderiing backend.
The interface now uses `Texture() TextureID` rather than `*gl.Texture`.
* TMXObject Width and Height is in pixels, and can be fractional. This has changed from an int to a float64.
* TMXTileset now uses a Spritesheet instead of a Texture. This helps keep track of the guid better and allows the gid to not start at zero and have skips in it, as well as for borders and spacing in the tile sheet.
* TMX Level's objects have all been rolled into Object rather than have separate things like "PolyLineObject". This is to be
consistent with the TMX format.
* The Shader interface now has a `SetCamera(*CameraSystem)` method. This method allows shaders to automatically update the camera system as it changes, such as between scenes or when the camera system is added.
* The domain engo.io has expired. Now use `github.com/EngoEngine/engo` as the import path, and the site can be located at
engoengine.github.io

## Roadmap to v1.1
A list of issues for v1.1 can be found [here](https://github.com/EngoEngine/engo/issues/552). There's always room
for improvement! Feel free to submit proposals, open issues, and let us know how we can improve!

## History

Engo, originally known as `Engi` was written by [ajhager](https://github.com/ajhager) as a general purpose Go game engine. With a desire to build it into an "ECS" game engine, it was forked to `github.com/paked/engi`. After passing through several iterations, it was decided that the project would be rebranded and rereleased as Engo on its own GitHub organization.

## Credits

Thank you to everyone who has worked on, or with `Engo`. None of this would be possible without you, and your help has been truly amazing.

- [ajhager](https://github.com/ajhager): Building the original `engi`, which engo was based off of
- [paked](https://github.com/paked): Adding ECS element, project maintenance and management
- [Newbrict](https://github.com/Newbrict): Font rendering, TMX support
- [EtienneBruines](https://github.com/EtienneBruines): Rewriting the OpenGL code, maintenance and helping redesign the API
- [otraore](https://github.com/otraore): Adding in GopherJS support, maintenance
- [Everyone else who has submitted PRs or issues over the years, to any iteration of the project](https://github.com/EngoEngine/engo/graphs/contributors)

These are 3rd party projects that have made `engo` possible.
- The original [engi](https://github.com/ajhager/engi) game engine which engo was based off of ([BSD license](https://github.com/ajhager/engi/blob/master/LICENSE))
- [Oto](https://github.com/hajimehoshi/oto), a low-level cross-platform library to play sound. The AudioSystem uses this and is based on
the audio package used in [Ebiten](https://github.com/hajimehoshi/ebiten).
