#version 410 core

layout (location = 0) in vec3 position;
layout (location = 1) in vec3 normal;
layout (location = 2) in vec4 color;
layout (location = 3) in vec2 texture;

uniform mat4 u_pvm;
uniform mat3 u_rot_stars;

out vec3 pos_sky;
out vec3 pos_stars;

void main()
{
    gl_Position = u_pvm * vec4(position, 1.0);
    pos_sky = position;
    pos_stars = u_rot_stars * position; 
}