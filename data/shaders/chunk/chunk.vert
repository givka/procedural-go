#version 410 core

layout (location = 0) in vec3 position;
layout (location = 1) in vec3 normal;
layout (location = 2) in vec4 color;
layout (location = 3) in vec2 texture;

uniform mat4 model;
uniform mat4 view;
uniform mat4 project;

uniform vec3 lightPos;  // only need one light for a basic example

out float RiverHeight;
out vec3 Normal;
out vec3 FragPos;
out vec3 LightPos;
out vec4 MatColor;
out vec2 TexCoord;
out float Height;

const float seaLevel = -1.9;

int getTexture()
{
    return 1;
}
void main()
{
    vec3 pos = position;
    if(-position.y < seaLevel) pos.y = 2.0;
    gl_Position = project * view * model * vec4(pos, 1.0);
    FragPos = gl_Position.xyz;
    LightPos = lightPos;

    mat3 normMatrix = mat3(transpose(inverse(view))) * mat3(transpose(inverse(model)));
    Normal = (transpose(inverse(model)) * vec4(normal, 1.0)).xyz;
    MatColor = color;
    TexCoord = texture;
    Height = -pos.y;
    RiverHeight = color.a;
}
