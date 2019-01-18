#version 410 core

layout (location = 0) in vec3 position;
layout (location = 1) in vec3 normal;
layout (location = 2) in vec4 color;
layout (location = 3) in vec2 texture;
layout (location = 4) in mat4 transform;

uniform mat4 view;
uniform mat4 project;
uniform mat4 model;

uniform vec3 lightPos;  // only need one light for a basic example

out vec3 Normal;
out vec3 FragPos;
out vec3 LightPos;
out vec4 MatColor;
out vec2 TexCoord;

void main()
{
    gl_Position = project * view * transform * model * vec4(position, 1.0);
    
    mat3 normMatrix = mat3(transpose(inverse(view))) * mat3(transpose(inverse(transform)));
    
    Normal = normMatrix * normal;
    FragPos = vec3(view * transform * vec4(position, 1.0));
    LightPos = vec3(view * vec4(lightPos, 1.0));
    MatColor = color;
    TexCoord = texture; 
}
