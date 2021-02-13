#version 450
#extension GL_ARB_separate_shader_objects : enable

layout(location = 0) in vec2 in_Position;
layout(location = 1) in vec2 in_TexCoords;
layout(location = 2) in vec4 in_Color;

layout(binding = 0) uniform UniformBufferObject {
  mat3 m;
} matrixProjView;

layout(location = 0) out vec4 var_Color;
layout(location = 1) out vec2 var_TexCoords;

void main() {
  var_Color = in_Color;
  var_TexCoords = in_TexCoords;

  vec3 matr = matrixProjView.m * vec3(in_Position, 1.0);
  gl_Position = vec4(matr.xy, 0, matr.z);
}
