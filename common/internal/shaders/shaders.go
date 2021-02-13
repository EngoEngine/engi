package shaders

//go:generate glslangvalidator -V blendmap/shader.frag -o blendmap/frag.spv
//go:generate glslangvalidator -V blendmap/shader.vert -o  blendmap/vert.spv
//go:generate glslangvalidator -V default/shader.frag -o default/frag.spv
//go:generate glslangvalidator -V default/shader.vert -o  default/vert.spv
//go:generate glslangvalidator -V legacy/shader.frag -o legacy/frag.spv
//go:generate glslangvalidator -V legacy/shader.vert -o  legacy/vert.spv
//go:generate glslangvalidator -V text/shader.frag -o text/frag.spv
//go:generate glslangvalidator -V text/shader.vert -o  text/vert.spv
//go:generate go-bindata -nocompress -pkg=shaders blendmap/frag.spv blendmap/vert.spv default/frag.spv default/vert.spv legacy/frag.spv legacy/vert.spv text/frag.spv text/vert.spv
//go:generate gofmt -s -w .
