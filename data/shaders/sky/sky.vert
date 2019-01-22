#version 410 core

layout (location = 0) in vec3 position;
layout (location = 1) in vec3 normal;
layout (location = 2) in vec4 color;
layout (location = 3) in vec2 texture;

uniform mat4 pvm;

out vec3 pos;

void main()
{
    gl_Position = pvm * vec4(position, 1.0);
    pos = position;
}